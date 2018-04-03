package push

import (
	"fmt"

	"github.com/yukongco/msgpush/comet/conf"
	"github.com/yukongco/msgpush/common/check"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
)

// IOS push is not div whether user is login
func PushToIOS(token string, message []byte, iosCertFile, topic string) error {
	notification := &apns2.Notification{}
	notification.DeviceToken = token
	//notification.Topic = "com.sideshow.Apns2"
	notification.Payload = message // See Payload section below

	// 获取 ios 的推送证书
	path := check.IOS_Cert_Prex + iosCertFile

	IOSCert, err := certificate.FromPemFile(path, conf.Conf.IosConf.Pwd)
	if err != nil {
		tmpStr := fmt.Sprintf("ios cert path=%v is err: %v", path, err)
		fmt.Println(tmpStr)
		return fmt.Errorf(tmpStr)
	}
	var client = apns2.NewClient(IOSCert)
	if conf.Conf.IosConf.Mode == check.Dev {
		client.Development()
	} else {
		notification.Topic = topic //production need topic, "<your-app-bundle-id>"
		client.Production()
	}

	res, err := client.Push(notification)
	if err != nil {
		fmt.Println("ios push err: ", err.Error())
		return err
	}

	if res.StatusCode != apns2.StatusSent {
		tmpStr := fmt.Sprintf("statusCode=%v, reason=%v", res.StatusCode, res.Reason)
		fmt.Printf(tmpStr)
		return fmt.Errorf(tmpStr)
	}

	return nil
}
