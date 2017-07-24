package lib

import (
	"errors"

	"github.com/graphql-go/graphql"
	"golang.org/x/net/context"
)

type IDFieldInterface interface {
	IdField() (*string, error)
}

func IDFetchFunction(obj interface{}, info graphql.ResolveInfo, ctx context.Context) (string, error) {
	field, ok := obj.(IDFieldInterface)
	if ok == false {
		return "", errors.New("Could not resolve the id")
	}
	id, err := field.IdField()
	return *id, err
}

type MutationPayload struct {
	ClientMutationID string
	Payload          interface{}
}

func AddFieldConfigMap(obj *graphql.Object, fields graphql.Fields) {
	for name, field := range fields {
		obj.AddFieldConfig(name, field)
	}
}
