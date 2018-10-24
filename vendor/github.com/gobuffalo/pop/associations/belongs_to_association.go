package associations

import (
	"fmt"
	"reflect"

	"github.com/gobuffalo/flect"
	"github.com/gobuffalo/pop/columns"
	"github.com/gobuffalo/pop/nulls"
)

// belongsToAssociation is the implementation for the belongs_to
// association type in a model.
type belongsToAssociation struct {
	ownerModel reflect.Value
	ownerType  reflect.Type
	ownerID    reflect.Value
	fkID       string
	primaryID  string
	ownedModel interface{}
	*associationSkipable
	*associationComposite

	primaryTableID string
}

func init() {
	associationBuilders["belongs_to"] = belongsToAssociationBuilder
}

func belongsToAssociationBuilder(p associationParams) (Association, error) {
	fval := p.modelValue.FieldByName(p.field.Name)
	primaryIDField := "ID"
	if p.popTags.Find("primary_id").Value != "" {
		primaryIDField = p.popTags.Find("primary_id").Value
	}

	ownerIDField := fmt.Sprintf("%s%s", p.field.Name, "ID")
	if p.popTags.Find("fk_id").Value != "" {
		ownerIDField = p.popTags.Find("fk_id").Value
	}

	if _, found := p.modelType.FieldByName(ownerIDField); !found {
		return nil, fmt.Errorf("there is no '%s' defined in model '%s'", ownerIDField, p.modelType.Name())
	}

	// Validates if ownerIDField is nil, this association will be skipped.
	var skipped bool
	f := p.modelValue.FieldByName(ownerIDField)
	if fieldIsNil(f) || isZero(f.Interface()) {
		skipped = true
	}
	//associated model
	ownerPrimaryTableField := "id"
	if primaryIDField != "ID" {
		ownerModel := reflect.Indirect(fval)
		ownerPrimaryField, found := ownerModel.Type().FieldByName(primaryIDField)
		if !found {
			return nil, fmt.Errorf("there is no primary field '%s' defined in model '%s'", primaryIDField, ownerModel.Type())
		}
		ownerPrimaryTags := columns.TagsFor(ownerPrimaryField)
		if dbField := ownerPrimaryTags.Find("db").Value; dbField == "" {
			ownerPrimaryTableField = flect.Underscore(ownerPrimaryField.Name) //autodetect without db tag
		} else {
			ownerPrimaryTableField = dbField
		}
	}

	return &belongsToAssociation{
		ownerModel: fval,
		ownerType:  fval.Type(),
		ownerID:    f,
		fkID:       ownerIDField,
		primaryID:  primaryIDField,
		ownedModel: p.model,
		associationSkipable: &associationSkipable{
			skipped: skipped,
		},
		associationComposite: &associationComposite{innerAssociations: p.innerAssociations},
		primaryTableID:       ownerPrimaryTableField,
	}, nil
}

func (b *belongsToAssociation) Kind() reflect.Kind {
	if b.ownerType.Kind() == reflect.Ptr {
		return b.ownerType.Elem().Kind()
	}
	return b.ownerType.Kind()
}

func (b *belongsToAssociation) Interface() interface{} {
	if b.ownerModel.Kind() == reflect.Ptr {
		val := reflect.New(b.ownerType.Elem())
		b.ownerModel.Set(val)
		return b.ownerModel.Interface()
	}
	return b.ownerModel.Addr().Interface()
}

// Constraint returns the content for a where clause, and the args
// needed to execute it.
func (b *belongsToAssociation) Constraint() (string, []interface{}) {
	return fmt.Sprintf("%s = ?", b.primaryTableID), []interface{}{b.ownerID.Interface()}
}

func (b *belongsToAssociation) BeforeInterface() interface{} {
	if !b.skipped {
		return nil
	}

	if b.ownerModel.Kind() == reflect.Ptr {
		return b.ownerModel.Interface()
	}

	currentVal := b.ownerModel.Interface()
	zeroVal := reflect.Zero(b.ownerModel.Type()).Interface()
	if reflect.DeepEqual(zeroVal, currentVal) {
		return nil
	}

	return b.ownerModel.Addr().Interface()
}

func (b *belongsToAssociation) BeforeSetup() error {
	ownerID := reflect.Indirect(reflect.ValueOf(b.ownerModel.Interface())).FieldByName("ID").Interface()
	if b.ownerID.CanSet() {
		if n := nulls.New(b.ownerID.Interface()); n != nil {
			b.ownerID.Set(reflect.ValueOf(n.Parse(ownerID)))
		} else {
			b.ownerID.Set(reflect.ValueOf(ownerID))
		}
		return nil
	}
	return fmt.Errorf("could not set '%s' to '%s'", ownerID, b.ownerID)
}
