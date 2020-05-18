package fieldfilter

import (
	"fmt"
	"reflect"
	"strings"
)

type Lanaguage = map[string]interface{}
type Lanaguages = map[string]Lanaguage

type Lang interface {
	Lang() Lanaguages
}

//根据过滤标签过滤单数据结构
// Tag
//		filter:"base"
//		filter:"self"
//		filter:"-" 禁显
//      filter:"*" 全显（结构数组不过滤子字段）
//      filter:"base:*;self:-;other" 全显（结构数组不过滤子字段）
func Filter(obj interface{}, t string, filters map[string]struct{}) interface{} {
	return FilterLang("", obj, t, filters)
}

func FilterLang(lang string, obj interface{}, t string, filters map[string]struct{}) interface{} {
	if len(filters) <= 0 {
		return map[string]interface{}{}
	}

	return filterobj(lang, obj, "", t, filters)
}

func FiltersToMap(filters string) map[string]struct{} {
	m := map[string]struct{}{}
	ss := strings.Split(filters, ",")
	for _, v := range ss {
		vs := strings.Split(v, ".")
		for i := 0; i < len(vs); i++ {
			m[strings.Join(vs[:i+1], ".")] = struct{}{}
		}
	}
	return m
}

func filterobj(lang string, obj interface{}, path string, ft string, filterM map[string]struct{}) interface{} {
	//	fmt.Println(obj, path, ft, filterM)
	if obj == nil {
		return nil
	}

	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		if !v.IsValid() {
			return nil
		}
		t = v.Type()
	}

	switch t.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64,
		reflect.String, reflect.Bool:
		return v.Interface()
	case reflect.Array, reflect.Slice:
		if v.IsNil() {
			return nil
		}
		if v.Len() > 0 {
			is := []interface{}{}
			switch v.Index(0).Type().Kind() {
			case reflect.Struct, reflect.Ptr:
				for i := 0; i < v.Len(); i++ {
					is = append(is, filterobj(lang, v.Index(i).Interface(), path, ft, filterM))
				}
				return is
			default:
				return v.Interface()
			}
		}
		return []int{}
	case reflect.Struct:
	case reflect.Map:
		return v.Interface()
	default:
		panic(fmt.Errorf("Unsupported type:%v", t.Kind()))
	}

	l, lok := v.Interface().(Lang)
	var cl Lanaguage
	if lok {
		cl = l.Lang()[lang]
	}

	//获取嵌套字段
	afield := map[string]struct{}{}
	pl := len(path)
	if pl > 0 {
		pl++
	}
	for k := range filterM {
		if len(k) >= pl {
			fields := strings.SplitN(k[pl:], ".", 3)
			if len(fields) > 1 {
				afield[fields[0]] = struct{}{}
			}
		}
	}
	//fmt.Println(t.Name(), path, "filter:afield:", filterM, afield)

	m := map[string]interface{}{}
	for i := 0; i < t.NumField(); i++ {
		fieldT := t.Field(i)
		fieldV := v.Field(i)
		if fieldT.Anonymous {
			amap := filterobj(lang, fieldV.Interface(), path, ft, filterM)
			for k, v := range amap.(map[string]interface{}) {
				m[k] = v
			}
			continue
		}
		jsonTag := fieldT.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = fieldT.Name
		}
		filterTag := fieldT.Tag.Get("filter")
		if filterTag != "" {
			// filter 不为空的时候判断是否显示
			ntag := ""
			tags := strings.Split(filterTag, ";")
			// 分割;
			for _, v := range tags {
				vs := strings.Split(v, ":")
				switch len(vs) {
				case 1:
					//单个
					switch vs[0] {
					case "-":
						ntag = "-"
					case "*":
						ntag = "*"
					default:
						if vs[0] == ft {
							ntag = ft
							break
						}
					}
				case 2:
					// 多个 k:v
					if vs[0] == ft {
						ntag = vs[1]
						break
					}
				}
			}
			filterTag = ntag
			if filterTag == "-" {
				//	fmt.Println(jsonTag, filterTag, "pass")
				continue
			} else if filterTag == "*" {
				//fmt.Println("f * cl", cl)
				if cl != nil {
					if lv, ok := cl[jsonTag]; ok {
						m[jsonTag] = lv
					} else {
						m[jsonTag] = fieldV.Interface()
					}
				} else {
					m[jsonTag] = fieldV.Interface()
				}
				continue
			} else if filterTag != ft {
				continue
			}
		}
		fpath := strings.TrimLeft(fmt.Sprintf("%v.%v", path, jsonTag), ".")
		//fmt.Println("filter:jsonTag:", afield, jsonTag, fpath)
		if _, ok := afield[jsonTag]; ok {
			m[jsonTag] = filterobj(lang, fieldV.Interface(), fpath, ft, filterM)
		} else if _, ok := filterM[fpath]; ok {
			//fmt.Println("filterM check:", fpath, fieldV.CanInterface())
			if fieldV.CanInterface() && jsonTag != "-" {
				if cl != nil {
					if lv, ok := cl[jsonTag]; ok {
						m[jsonTag] = lv
					} else {
						m[jsonTag] = filterobj(lang, fieldV.Interface(), fpath, ft, filterM)
					}
				} else {
					m[jsonTag] = filterobj(lang, fieldV.Interface(), fpath, ft, filterM)
				}
			}
		}
	}

	return m
}
