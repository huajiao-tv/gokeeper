package models

func Init(keeperAddr string) error {
	InitClient(keeperAddr)
	return nil
}
