package controlloop

import "errors"

var KetNotExist = errors.New("key does't exist")
var AlreadyUpdated = errors.New("resource already updated")
