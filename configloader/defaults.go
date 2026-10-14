package configloader

import (
	"reflect"
	"strings"

	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

// decodeConfig uses mapstructure for every input conversion. Nested compound
// fields get their own decoder scope so their defaults policy follows the field,
// including when the decoder allocates pointer, slice or map children.
// Defaults precede the source overlay; decoded zero values are never defaulted.
func decodeConfig(input interface{}, output reflect.Value, allowDefaults bool) error {
	initialized := make(map[interface{}]bool)
	markExistingStructs(output, initialized, make(map[interface{}]bool))
	return decodeConfigScope(input, output, allowDefaults, initialized)
}

func decodeConfigScope(input interface{}, output reflect.Value, allowDefaults bool, initialized map[interface{}]bool) error {
	scopes := make(map[interface{}]bool)
	replaced := make(map[interface{}]bool)
	rootAddress := valueAddress(output)
	unscopedDefaults := allowDefaults
	hook := func(from, to reflect.Value) (interface{}, error) {
		address := valueAddress(to)
		// Replace a provided container once. Weak slice-to-map conversion revisits
		// the same destination while merging its elements; it must not reset it again.
		if address != nil && !replaced[address] && (to.Kind() == reflect.Slice || to.Kind() == reflect.Map) {
			to.Set(reflect.Zero(to.Type()))
			replaced[address] = true
		}
		if allowed, ok := scopes[address]; address != nil && ok {
			delete(scopes, address)
			prepared := reflect.New(to.Type()).Elem()
			prepared.Set(to)
			if initialized[address] {
				markExistingStructs(prepared, initialized, make(map[interface{}]bool))
			}
			if err := decodeConfigScope(from.Interface(), prepared, allowed, initialized); err != nil {
				return nil, err
			}
			// The native decoder copies prepared structs, including their supplied zeros.
			// Containers are still decoded natively; their children are now typed values.
			return preparedInput(prepared, to), nil
		}
		if to.Kind() == reflect.Struct && address != nil {
			data := reflect.Indirect(from)
			if data.IsValid() && data.Type() == to.Type() {
				// Native direct assignment treats every field in a typed struct as supplied,
				// even fields whose value is zero or whose source tag says omitempty.
				return from.Interface(), nil
			}
			if address != rootAddress {
				// Decoder-created elements need their own struct scope. Besides isolating
				// same-type siblings, this contains the native remain-field decoding,
				// which bypasses the callback for the enclosing map field.
				prepared := reflect.New(to.Type()).Elem()
				prepared.Set(to)
				if initialized[address] {
					markExistingStructs(prepared, initialized, make(map[interface{}]bool))
				}
				if err := decodeConfigScope(from.Interface(), prepared, unscopedDefaults, initialized); err != nil {
					return nil, err
				}
				return prepared.Interface(), nil
			}
			if allowDefaults && !initialized[address] && to.IsZero() {
				if err := defaults.Set(to.Addr().Interface()); err != nil {
					return nil, errors.Wrap(err, "failed to set nested defaults")
				}
				markExistingStructs(to, initialized, make(map[interface{}]bool))
			}
			unscopedDefaults = registerFieldScopes(to, allowDefaults, scopes)
		}
		return from.Interface(), nil
	}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result: output.Addr().Interface(), WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(), mapstructure.StringToSliceHookFunc(","), hook),
	})
	if err != nil {
		return err
	}
	return decoder.Decode(input)
}

// Replay the completed value, not its pointer allocation wrappers: native
// struct/scalar/container conversion only dereferences input once. Containers
// need the same treatment for pointer elements; complete structs remain typed
// so their zero fields are copied. The outer decoder still allocates/reuses
// every pointer level in its actual destination.
func preparedInput(prepared, destination reflect.Value) interface{} {
	for prepared.Kind() == reflect.Ptr {
		if prepared.IsNil() {
			return nil
		}
		prepared = prepared.Elem()
	}
	for destination.Kind() == reflect.Ptr && !destination.IsNil() {
		destination = destination.Elem()
	}
	if prepared.Kind() == reflect.Interface {
		// Preserve the complete dynamic value (including typed nil/deep pointers).
		// A *interface input lets native decodeBasic copy that interface directly.
		replay := reflect.New(prepared.Type())
		replay.Elem().Set(prepared)
		if destination.Kind() == reflect.Interface && destination.CanSet() {
			destination.Set(reflect.Zero(destination.Type()))
		}
		return replay.Interface()
	}
	switch prepared.Kind() {
	case reflect.Slice, reflect.Array:
		if prepared.Kind() == reflect.Slice && prepared.IsNil() {
			return prepared.Interface()
		}
		replay := make([]interface{}, prepared.Len())
		for i := range replay {
			var child reflect.Value
			if (destination.Kind() == reflect.Slice || destination.Kind() == reflect.Array) && i < destination.Len() {
				child = destination.Index(i)
			}
			replay[i] = preparedInput(prepared.Index(i), child)
		}
		return replay
	case reflect.Map:
		if prepared.IsNil() {
			return prepared.Interface()
		}
		replay := reflect.MakeMapWithSize(reflect.MapOf(prepared.Type().Key(), reflect.TypeOf((*interface{})(nil)).Elem()), prepared.Len())
		for _, key := range prepared.MapKeys() {
			var child reflect.Value
			if destination.Kind() == reflect.Map {
				child = destination.MapIndex(key)
			}
			data := preparedInput(prepared.MapIndex(key), child)
			value := reflect.New(replay.Type().Elem()).Elem()
			if data != nil {
				value.Set(reflect.ValueOf(data))
			}
			replay.SetMapIndex(key, value)
		}
		return replay.Interface()
	}
	return prepared.Interface()
}

// Addresses retain their Go pointers, keeping decoder temporaries alive. Policy
// keys therefore cannot be confused by allocation/address reuse during a parse.
func valueAddress(value reflect.Value) interface{} {
	if value.CanAddr() && value.Addr().CanInterface() {
		return value.Addr().Interface()
	}
	return nil
}

func registerFieldScopes(value reflect.Value, allowed bool, scopes map[interface{}]bool) bool {
	type structScope struct {
		value   reflect.Value
		allowed bool
	}
	queue := []structScope{{value, allowed}}
	remainingAllowed := allowed
	// Match the decoder's breadth-first squash traversal and last remain field.
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for i := 0; i < current.value.NumField(); i++ {
			field := current.value.Type().Field(i)
			child := current.value.Field(i)
			if !child.CanSet() {
				continue
			}
			childAllowed := current.allowed && field.Tag.Get("default") != "-"
			mode := ""
			for _, tag := range strings.Split(field.Tag.Get("mapstructure"), ",")[1:] {
				if tag == "squash" || tag == "remain" {
					mode = tag
					break
				}
			}
			if mode == "squash" {
				nested := child
				if nested.Kind() == reflect.Ptr && !nested.IsNil() {
					nested = nested.Elem()
				}
				if nested.Kind() == reflect.Struct {
					queue = append(queue, structScope{nested, childAllowed})
				}
				continue
			}
			if mode == "remain" {
				remainingAllowed = childAllowed
				continue
			}
			// The native struct decoder uses an existing *struct's element as
			// the field destination; nil pointers are decoded through their slot.
			destination := child
			if child.Kind() == reflect.Ptr && child.Elem().Kind() == reflect.Struct {
				destination = child.Elem()
			}
			if address := valueAddress(destination); address != nil {
				switch destination.Kind() {
				case reflect.Struct, reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Array, reflect.Map:
					scopes[address] = childAllowed
				}
			}
		}
	}
	return remainingAllowed
}

// The root defaults pass has already initialized existing objects. Remember them
// before decoding, so nested decoders only initialize newly allocated structs.
func markExistingStructs(value reflect.Value, initialized, visited map[interface{}]bool) {
	if !value.IsValid() {
		return
	}
	if address := valueAddress(value); address != nil {
		if visited[address] {
			return
		}
		visited[address] = true
		if value.Kind() == reflect.Struct {
			initialized[address] = true
		}
	}
	switch value.Kind() {
	case reflect.Ptr:
		if !value.IsNil() {
			markExistingStructs(value.Elem(), initialized, visited)
		}
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			if value.Field(i).CanInterface() && value.Type().Field(i).Tag.Get("default") != "-" {
				markExistingStructs(value.Field(i), initialized, visited)
			}
		}
	case reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			markExistingStructs(value.Index(i), initialized, visited)
		}
	case reflect.Map:
		for _, key := range value.MapKeys() {
			markExistingStructs(value.MapIndex(key), initialized, visited)
		}
	}
}
