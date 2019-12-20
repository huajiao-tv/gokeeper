package model

// ObjContainer is function to update data
type ObjContainer interface {
	Update(name string, st interface{}) error
	GetStructs() map[string]interface{}
}
