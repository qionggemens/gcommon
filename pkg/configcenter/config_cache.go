/**
 * Create Time:2023/5/19
 * User: luchao
 * Email: lcmusic1994@gmail.com
 */

package configcenter

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"gopkg.in/yaml.v2"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var configClient config_client.IConfigClient
var configMap = make(map[string]interface{}, 0)

func LoadYamlConfig(namespaceId string, dataId string) error {
	err := initConfigClient(namespaceId, dataId)
	if err != nil {
		return err
	}
	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: dataId,
		Group:  "DEFAULT_GROUP",
	})
	if err != nil {
		return err
	}
	if content == "" {
		return nil
	}

	confMap := make(map[string]interface{}, 0)
	err = yaml.Unmarshal([]byte(content), confMap)
	if err != nil {
		return err
	}
	buildFlattenedMap(configMap, confMap, "")
	glog.Infoln("-------------------------- nacos config --------------------------")
	for k, v := range configMap {
		glog.Infof("---  %s: %v", k, v)
	}
	return nil
}

//
// getClientConfig
//  @Description: 获取客户端配置
//  @param namespaceId
//  @param moduleId
//  @return *constant.ClientConfig
//
func getClientConfig(namespaceId string, dataId string) *constant.ClientConfig {
	logDir := fmt.Sprintf("/tmp/nacos/log/%s/%s", namespaceId, dataId)
	cacheDir := fmt.Sprintf("/tmp/nacos/cache/%s/%s", namespaceId, dataId)
	return constant.NewClientConfig(
		constant.WithNamespaceId(namespaceId),
		constant.WithTimeoutMs(5000),
		constant.WithNotLoadCacheAtStart(true),
		constant.WithLogDir(logDir),
		constant.WithCacheDir(cacheDir),
		constant.WithLogLevel("debug"),
	)
}

//
// getServerConfig
//  @Description: 获取服务端配置
//  @return []constant.ServerConfig
//
func getServerConfig() []constant.ServerConfig {
	nacosIp := os.Getenv("NACOS_IP")
	nacosPort := os.Getenv("NACOS_PORT")
	if nacosIp == "" {
		nacosIp = "qionggemen.nacos.net"
	}
	if nacosPort == "" {
		nacosPort = "8848"
	}
	port, _ := strconv.ParseUint(nacosPort, 10, 64)
	return []constant.ServerConfig{
		*constant.NewServerConfig(nacosIp, port, constant.WithContextPath("/nacos")),
	}
}

//
// initConfigClient
//  @Description: 初始化nacos客户端
//  @param namespaceId
//  @param moduleId
//  @return error
//
func initConfigClient(namespaceId string, dataId string) error {
	cc := getClientConfig(namespaceId, dataId)
	sc := getServerConfig()
	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		return err
	}
	configClient = client
	return nil
}

//
// buildFlattenedMap
//  @Description: Spring 原生转换
//  @param result
//  @param source
//  @param path
//
func buildFlattenedMap(result map[string]interface{}, source map[string]interface{}, path string) {
	for k, v := range source {
		if len(path) != 0 && path != "" {
			if strings.HasPrefix(k, "[") {
				k = path + k
			} else {
				k = path + "." + k
			}
		}
		vn := reflect.TypeOf(v).Kind()
		if vn == reflect.String {
			result[k] = v
		} else if vn == reflect.Map {
			value := v.(map[interface{}]interface{})
			son := make(map[string]interface{}, 0)
			for mk, mv := range value {
				son[mk.(string)] = mv
			}
			buildFlattenedMap(result, son, k)
		} else if vn == reflect.Array || vn == reflect.Slice {
			value := v.([]interface{})
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
			if v != nil {
				result[k] = v
			} else {
				result[k] = ""
			}
		}
	}
}
