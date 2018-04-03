package check

import (
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"
)

// 得到一个标准的 id, 如： 00347135-859b-4561-aea0-cb03d71bb47a
func GetStandId() string {
	kinds := [][]int{[]int{10, 48}, []int{26, 97}, []int{26, 65}}
	res := make([]byte, 36)

	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 36; i++ {
		// random ikind
		ikind := rand.Intn(2)
		scope := kinds[ikind][0]
		base := kinds[ikind][1]
		rand.Intn(2)

		if i == 8 || i == 13 || i == 18 || i == 23 {
			res[i] = uint8(45)
		} else {
			res[i] = uint8(base + rand.Intn(scope))
		}
	}

	return string(res)
}

// get local ip
func GetLoacalIp() (string, error) {
	conn, err := net.Dial("udp", "github.com:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localIp := strings.Split(conn.LocalAddr().String(), ":")[0]
	if localIp == "" {
		return "", fmt.Errorf("loacal ip is nil")
	}

	return localIp, nil
}
