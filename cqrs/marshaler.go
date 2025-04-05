package cqrs

import (
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/dmitrymomot/gokit/utils"
	"github.com/google/uuid"
)

// marshaler is a global marshaler for the service.
var marshaler = cqrs.JSONMarshaler{
	NewUUID:      func() string { return uuid.NewString() },
	GenerateName: func(v any) string { return utils.GetNameFromStruct(v, utils.StructName) },
}
