package structencoder

import (
	"github.com/anexia-it/go-cloudlog"
	"gopkg.in/anexia-it/go-structmapper.v1"
)

// DefaultTagName defines the default tag name to use
const DefaultTagName = "cloudlog"

var _ cloudlog.EventEncoder = (*StructEncoder)(nil)

type StructEncoder struct {
	mapper *structmapper.Mapper
}

func (e *StructEncoder) EncodeEvent(event interface{}) (m map[string]interface{}, err error) {
	return e.mapper.ToMap(event)
}

// NewStructEncoder returns a new encoder that supports structs
func NewStructEncoder(options ...structmapper.Option) (*StructEncoder, error) {
	mapper, err := structmapper.NewMapper(append([]structmapper.Option{
		structmapper.OptionTagName(DefaultTagName)}, options...)...)
	if err != nil {
		return nil, err
	}

	return &StructEncoder{
		mapper: mapper,
	}, nil
}
