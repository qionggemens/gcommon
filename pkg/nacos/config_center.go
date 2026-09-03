/**
 * Create Time:2024/1/12
 * User: luchao
 * Email: lcmusic1994@gmail.com
 */

package nacos

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/qionggemens/gcommon/pkg/glog"
	"gopkg.in/yaml.v2"
)

var configMap = make(map[string]interface{}, 0)

func LoadYamlConfig(namespaceId string, dataId string) error {
	cc := getClientConfig(namespaceId, dataId)
	sc := getServerConfig()
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		glog.Errorf("LoadYamlConfig fail - namespaceId:%s, dataId:%s, msg:%s", namespaceId, dataId, err.Error())
		return errors.New("LoadYamlConfig fail")
	}
	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  "DEFAULT_GROUP",
	})
	if err != nil {
		glog.Errorf("LoadYamlConfig fail - namespaceId:%s, dataId:%s, msg:%s", namespaceId, dataId, err.Error())
		return errors.New("LoadYamlConfig fail")
	}
	if content == "" {
		glog.Warningf("LoadYamlConfig finish - namespaceId:%s, dataId:%s, msg:config is empty", namespaceId, dataId)
		return nil
	}

	confMap := make(map[string]interface{}, 0)
	err = yaml.Unmarshal([]byte(content), confMap)
	if err != nil {
		return err
	}
	buildFlattenedMap(configMap, confMap, "")
	glog.Infof("-------------------------- nacos config --------------------------")
	for k, v := range configMap {
		glog.Infof("---  %s: %v", k, v)
	}
	glog.Infof("LoadYamlConfig finish - namespaceId:%s, dataId:%s", namespaceId, dataId)
	return nil
}

// buildFlattenedMap 对齐 Spring 风格扁平化；nil 值安全处理。
func buildFlattenedMap(result map[string]interface{}, source map[string]interface{}, path string) {
	for k, v := range source {
		if path != "" {
			if strings.HasPrefix(k, "[") {
				k = path + k
			} else {
				k = path + "." + k
			}
		}
		if v == nil {
			result[k] = ""
			continue
		}
		vn := reflect.TypeOf(v).Kind()
		if vn == reflect.String {
			result[k] = v
		} else if vn == reflect.Map {
			value, ok := v.(map[interface{}]interface{})
			if !ok {
				result[k] = fmt.Sprint(v)
				continue
			}
			son := make(map[string]interface{}, 0)
			for mk, mv := range value {
				son[fmt.Sprint(mk)] = mv
			}
			buildFlattenedMap(result, son, k)
		} else if vn == reflect.Array || vn == reflect.Slice {
			value, ok := v.([]interface{})
			if !ok {
				result[k] = fmt.Sprint(v)
				continue
			}
			if len(value) == 0 {
				result[k] = ""
			} else {
				for i, j := range value {
					m := make(map[string]interface{}, 0)
					m["["+strconv.FormatInt(int64(i), 10)+"]"] = j
					buildFlattenedMap(result, m, k)
				}
			}
		} else {
			result[k] = v
		}
	}
}

func GetBool(key string, defValue bool) bool {
	value, ok := configMap[key]
	if !ok {
		return defValue
	}
	switch v := value.(type) {
	case bool:
		return v
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "true", "yes", "y", "on":
			return true
		case "0", "false", "no", "n", "off":
			return false
		}
	}
	return defValue
}

func GetString(key string, defValue string) string {
	value, isOk := configMap[key]
	if !isOk {
		return defValue
	}
	if val, isOk := value.(string); isOk {
		return val
	}
	return fmt.Sprint(value)
}

func GetStrList(key string) []string {
	result := make([]string, 0)
	i := int64(0)
	for {
		k := fmt.Sprintf("%s[%s]", key, strconv.FormatInt(i, 10))
		value, isOk := configMap[k]
		if !isOk {
			break
		}
		val, isOk := value.(string)
		if isOk {
			result = append(result, val)
		} else {
			break
		}
		i++
	}
	return result
}

func GetInt(key string, defValue int) int {
	n, ok := toInt64(configMap[key])
	if !ok {
		return defValue
	}
	return int(n)
}

func GetInt32(key string, defValue int32) int32 {
	n, ok := toInt64(configMap[key])
	if !ok {
		return defValue
	}
	return int32(n)
}

func GetInt64(key string, defValue int64) int64 {
	n, ok := toInt64(configMap[key])
	if !ok {
		return defValue
	}
	return n
}

// toInt64 兼容 yaml.v2 常见数字类型（int / int64 / float64 / string）
func toInt64(value interface{}) (int64, bool) {
	if value == nil {
		return 0, false
	}
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	case float32:
		return int64(v), true
	case float64:
		return int64(v), true
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil {
			return 0, false
		}
		return n, true
	default:
		return 0, false
	}
}
