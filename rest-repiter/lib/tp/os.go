package tp

import "io/ioutil"

func OpenReadFile(filePath string)(dat []byte,err error)  {
	dat, err = ioutil.ReadFile(filePath)
	return
}