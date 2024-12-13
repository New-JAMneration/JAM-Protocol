package sum

import "github.com/New-JAMneration/JAM-Protocol/logger"

func Sum(a, b int) {
	logger.Infof("The sum of %d and %d is %d\n", a, b, a+b)
}
