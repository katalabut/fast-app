package configloader

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/mitchellh/mapstructure"
)

// All input shapes are interpreted by the same native decoder, including errors.
// A suppressed field must preserve native pointer depth, nilness and dynamic type.
func TestPointerDepthNativeParity(t *testing.T) {
	nodeType := reflect.TypeOf(policyChild{})
	bases := []reflect.Type{nodeType, reflect.TypeOf(0), reflect.TypeOf(false), reflect.TypeOf(""), reflect.TypeOf(time.Duration(0)), reflect.TypeOf([]policyChild{}), reflect.TypeOf(map[string]policyChild{}), reflect.TypeOf([1]policyChild{}), reflect.TypeOf((*interface{})(nil)).Elem()}
	ifacePayloads := []interface{}{policyChild{Name: "provided"}, new(policyChild), (*policyChild)(nil), new(*policyChild)}
	inputs := []interface{}{nil, (*policyChild)(nil), map[string]interface{}(nil), map[string]interface{}{}, map[string]interface{}{"name": "provided", "port": 0, "enabled": false}, policyChild{Name: "provided"}, &policyChild{Name: "provided"}, []interface{}{map[string]interface{}{"name": "provided", "port": 0, "enabled": false}, nil}, map[string]interface{}{"main": map[string]interface{}{"name": "provided", "port": 0, "enabled": false}}, 0, false, "0", "0s", ""}
	for _, payload := range ifacePayloads {
		box := payload
		inputs = append(inputs, &box)
	}
	for _, base := range bases {
		for depth := 1; depth <= 4; depth++ {
			typ := base
			for level := 0; level < depth; level++ {
				typ = reflect.PointerTo(typ)
			}
			cfgType := reflect.StructOf([]reflect.StructField{{Name: "Value", Type: typ, Tag: `default:"-"`}})
			for state := 0; state < 3; state++ {
				for index, input := range inputs {
					t.Run(fmt.Sprintf("%s/depth%d/state%d/input%d", base, depth, state, index), func(t *testing.T) {
						initial := func() reflect.Value {
							result := reflect.New(cfgType)
							if state > 0 {
								slot := result.Elem().Field(0)
								slot.Set(reflect.New(slot.Type().Elem()))
								if state == 2 {
									for slot.Kind() == reflect.Ptr {
										if slot.IsNil() {
											slot.Set(reflect.New(slot.Type().Elem()))
										}
										slot = slot.Elem()
									}
									if slot.Type() == nodeType {
										slot.Set(reflect.ValueOf(policyChild{Name: "retained", Port: 11, Enabled: true}))
									}
									if slot.Kind() == reflect.Interface {
										slot.Set(reflect.ValueOf(&policyChild{Name: "retained", Port: 11, Enabled: true}))
									}
								}
							}
							return result
						}
						actual := initial()
						expected := initial()
						p, _ := NewParser()
						_ = p.SetSource(formSource{"value": input})
						err := p.Parse(actual.Interface())
						decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{Result: expected.Interface(), WeaklyTypedInput: true, DecodeHook: mapstructure.ComposeDecodeHookFunc(mapstructure.StringToTimeDurationHookFunc(), mapstructure.StringToSliceHookFunc(","))})
						nativeErr := decoder.Decode(p.viper.AllSettings())
						if (err == nil) != (nativeErr == nil) {
							t.Fatalf("error mismatch actual=%v native=%v", err, nativeErr)
						}
						if nativeErr == nil && !reflect.DeepEqual(actual.Interface(), expected.Interface()) {
							t.Fatalf("result differs from native: actual=%#v native=%#v", actual.Elem().Field(0).Interface(), expected.Elem().Field(0).Interface())
						}
					})
				}
			}
		}
	}
}

func TestPointerDepthDefaultsAndWrappers(t *testing.T) {
	nodeType := reflect.TypeOf(policyChild{})
	wrappers := []reflect.Type{nodeType, reflect.TypeOf([]policyChild{}), reflect.TypeOf(map[string]policyChild{}), reflect.TypeOf([]**policyChild{}), reflect.TypeOf(map[string]***policyChild{})}
	for _, base := range wrappers {
		for depth := 0; depth <= 4; depth++ {
			typ := base
			for level := 0; level < depth; level++ {
				typ = reflect.PointerTo(typ)
			}
			cfgType := reflect.StructOf([]reflect.StructField{{Name: "Value", Type: typ}, {Name: "Skipped", Type: typ, Tag: `default:"-"`}})
			for _, explicit := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/depth%d/explicit%v", base, depth, explicit), func(t *testing.T) {
					leaf := map[string]interface{}{"name": "provided"}
					if explicit {
						leaf["port"] = 0
						leaf["enabled"] = false
					}
					var input interface{} = leaf
					switch base.Kind() {
					case reflect.Slice:
						input = []interface{}{leaf, nil}
					case reflect.Map:
						input = map[string]interface{}{"main": leaf}
					}
					cfg := reflect.New(cfgType)
					p, _ := NewParser()
					_ = p.SetSource(formSource{"value": input, "skipped": input})
					if err := p.Parse(cfg.Interface()); err != nil {
						t.Fatal(err)
					}
					node := formNode(derefPointers(cfg.Elem().Field(0)))
					skip := formNode(derefPointers(cfg.Elem().Field(1)))
					port, enabled := 7, true
					if explicit {
						port, enabled = 0, false
					}
					if node.Port != port || node.Enabled != enabled || node.Timeout != 2*time.Second || skip.Port != 0 || skip.Enabled || skip.Timeout != 0 {
						t.Fatalf("depth policy ordinary=%+v skipped=%+v", node, skip)
					}
				})
			}
		}
	}
}
func derefPointers(value reflect.Value) reflect.Value {
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	return value
}

type namedPolicyPointer *policyChild
type namedOuterPointer *namedPolicyPointer

func TestNamedPointerTargets(t *testing.T) {
	type Cfg struct {
		Value   namedOuterPointer
		Skipped namedOuterPointer `default:"-"`
		List    *[]**policyChild
	}
	p, _ := NewParser()
	leaf := map[string]interface{}{"name": "provided", "port": 0, "enabled": false}
	_ = p.SetSource(formSource{"value": leaf, "skipped": leaf, "list": []interface{}{leaf, nil}})
	var cfg Cfg
	if err := p.Parse(&cfg); err != nil {
		t.Fatal(err)
	}
	if (**cfg.Value).Port != 0 || (**cfg.Value).Enabled || (**cfg.Value).Timeout != 2*time.Second || (**cfg.Skipped).Timeout != 0 || (**(*cfg.List)[0]).Timeout != 2*time.Second || (*cfg.List)[1] != nil {
		t.Fatal("named/collection pointer policy changed")
	}
}
