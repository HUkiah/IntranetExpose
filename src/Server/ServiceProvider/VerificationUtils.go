package ServiceProvider

import (
	"fmt"
	"strconv"
)

type Port struct {
	port int
}

func (pt *Port) Error() string {

	strFormat := `you port is %d , Notice that the range is between 1-65536. `
	return fmt.Sprintf(strFormat, pt.port)
}

func Atoi(str string) (errorMsg string) {

	if port, _ := strconv.Atoi(str); !(port >= 1 && port < 65536) {

		dData := Port{
			port: port,
		}

		errorMsg = dData.Error()
		return
	} else {
		return ""
	}
}
