package filters

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"goxy/internal/common"
	"goxy/internal/proxy/http/wrapper"
	"regexp"
)

func NewTransformVolgaCTF() Rule {
	return TransformVolgaCTF{regex: regexp.MustCompile("VolgaCTF{([^}]+)}")}
}

type TransformVolgaCTF struct {
	regex *regexp.Regexp
}

func (t TransformVolgaCTF) Apply(_ *common.ProxyContext, e wrapper.Entity) (bool, error) {
	// TODO: apply to all fields (headers, etc)
	body, err := e.GetBody()
	if err != nil {
		return false, fmt.Errorf("get body: %w", err)
	}
	logrus.Debugf("Got body: %s", string(body))
	matches := t.regex.FindAllSubmatch(body, -1)
	logrus.Debugf("Got matches: %+v", matches)
	for _, match := range matches {
		body = bytes.ReplaceAll(body, match[1], obfuscateMatch(match[1]))
	}
	logrus.Debugf("New body: %s", string(body))
	e.SetBody(body)
	return len(matches) > 0, nil
}

func (t TransformVolgaCTF) String() string {
	return "transform-VolgaCTF"
}

func obfuscateMatch(match []byte) []byte {
	result := make([]byte, len(match))
	for i := 0; i+1 < len(result); i += 2 {
		result[i], result[i+1] = match[i+1], match[i]
	}
	return result
}
