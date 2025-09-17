package dataservice

import (
	"acetek-mes/model"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/yxcloud1/go-comm/db"
)

func FindDeviceByIP(addr string, port string) (string, string, string, int) {
	if res, err := db.DB().ExecuteQuery("exec sp_lims_query_device_by_ip @addr = ? , @type = ? ", addr, port); err != nil {
		return addr, port, "", 1000
	} else {
		if len(res) > 0 {
			res1 := addr
			res2 := port
			res4 := ""
			res3 := 1500
			if v, ok := res[0]["type"]; ok {
				res1 = fmt.Sprintf("%v", v)
			}
			if v, ok := res[0]["addr"]; ok {
				res2 = fmt.Sprintf("%v", v)
			}
			if v, ok := res[0]["delay"]; ok {
				if t, err := strconv.Atoi(fmt.Sprintf("%v", v)); err != nil {
					res3 = 1500
				} else {
					res3 = t
				}
				if res3 < 100 {
					res3 = 100
				}
			}
			if v, ok := res[0]["end_flag"]; ok {
				res4 = fmt.Sprintf("%v", v)
			}
			return res1, res2, res4, res3
		} else {
			return addr, port, "", 1500
		}
	}
}

func bytesToHex(byts []byte) string {
	var sb strings.Builder
	for i, b := range byts {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(fmt.Sprintf("%02X", b))
	}
	return sb.String()
}

func SaveDcLog(url string, client_ip string, deviceType string, deviceId string, request string, body []byte) (int, error) {
	log := &model.LimsDcRequestLog{
		RequestUrl: url,
		DeviceType: deviceType,
		DeviceID:   deviceId,
		Request:    request,
		RawData:    bytesToHex(body),
		ClientIP:   client_ip,
	}
	if tx := db.DB().Conn().Save(log); tx.Error != nil {
		return 0, tx.Error
	} else {
		return log.ID, nil
	}
}

func SaveReciveeData(rawid int, deviceType string, deviceId string, data string, values []string) map[string]any {
	command := `EXEC sp_lims_save_dc_data @deviceType= ? ,@deviceID= ? ,@sampleID= ? ,@rawID= ? ,@rawData= ? , @item_codes= ? ,
								     @item_value1= ? , @item_value2= ? , @item_value3= ? , @item_value4= ? , @item_value5= ? ,
									 @item_value6= ? , @item_value7= ? , @item_value8= ? , @item_value9= ? , @item_value10= ? ,
									 @item_value11= ? , @item_value12= ? , @item_value13= ? , @item_value14= ? , @item_value15= ? `
	var params []interface{}
	params = append(params, deviceType, deviceId, "", rawid, data, "")
	for i := 0; i < 15; i++ {
		if len(values) > i {
			params = append(params, values[i])
		} else {
			params = append(params, nil)
		}
	}
	err := db.DB().ExecuteSQL(command, params...)
	if err != nil {
		log.Println(err)
	}
	return map[string]any{
		"type": deviceType,
		"id":   deviceId,
	}
}

