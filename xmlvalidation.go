package ginxml

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/libxml2"
	"github.com/lestrrat-go/libxml2/xsd"
	"go.uber.org/zap"
)

type XmlValidator struct {
	Loger   *zap.Logger
	Xsdfile string
	xsd     *xsd.Schema
}

func InitXMLValidator(file string, logger *zap.Logger) (*XmlValidator, error) {
	service := &XmlValidator{
		Xsdfile: file,
		Loger:   logger,
	}

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		logger.Error("load xsd file failed", zap.String("file", file), zap.Error(err))
		return nil, err
	}

	s, err := xsd.Parse(buf)
	if err != nil {
		logger.Error("parse xsd file failed", zap.String("file", file), zap.Error(err))
		return nil, err
	}

	service.xsd = s

	logger.Info("load xsd file done.", zap.String("file", file))

	return service, nil
}

func (xv *XmlValidator) Validate(data []byte) error {
	if xv.xsd == nil {
		return fmt.Errorf("xsd is not ready yet")
	}
	//load xml
	d, err := libxml2.Parse(data)
	if err != nil {
		xv.Loger.Error("parse xml failed.", zap.Error(err))
		return err
	}
	defer d.Free()
	if err = xv.xsd.Validate(d); err != nil {
		for _, e := range err.(xsd.SchemaValidationError).Errors() {
			xv.Loger.Error("validate xml return error, ", zap.String("error", e.Error()))
		}
		return err
	}
	return nil
}

func (xv *XmlValidator) Middleware(c *gin.Context) {
	data := make([]byte, 0)

	if c.Request.Body != nil {
		data, _ = ioutil.ReadAll(c.Request.Body)
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	}
	xv.Validate(data)
	// if err != nil {
	// 	// panic(err)
	// }
	c.Next()
}
