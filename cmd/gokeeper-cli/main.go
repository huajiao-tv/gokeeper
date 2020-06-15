package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/huajiao-tv/gokeeper/model"
	"github.com/huajiao-tv/gokeeper/server/conf"
)

var (
	libraryMap = map[string]string{"time.Duration": "time"} //template libraries
)

var (
	fileInput  string
	fileOutput string

	domain string
	keeper string

	confSource string
)

func init() {
	flag.StringVar(&fileInput, "i", "", "Input directory")
	flag.StringVar(&fileOutput, "o", "./data", "Output directory")
	flag.StringVar(&domain, "d", "", "Domain name")
	flag.StringVar(&keeper, "k", "", "Keeper address")
	flag.StringVar(&confSource, "s", "l", "Conf source")

	flag.Parse()
}

func main() {
	var confManager *conf.ConfManager
	var err error

	switch confSource {
	//from local
	case "l":
		confManager, err = conf.NewConfManagerFromLocal(fileInput, "")
	//from network(keeper)
	case "n":
		confManager, err = getConfManagerFromNetwork(keeper, domain)
	}
	if err != nil {
		renderError(err)
	}

	structDatas := []model.StructData{}
	jsonFields := map[string]JF{}
	for structName, ckdMap := range confManager.GetKeyData() {
		sd := model.NewStructData(structName, 0, convertCfd(ckdMap))
		if len(sd.Data) == 0 {
			continue
		}
		for _, d := range sd.Data {
			if !d.IsJson {
				continue
			}
			if _, ok := jsonFields[d.Type]; ok {
				continue
			}
			jsonFields[d.Type] = JF{Type: d.Type, Value: d.RawValue}
		}

		sd.Libraries = getLibraries(sd)
		structDatas = append(structDatas, sd)
	}

	err = json2Struct("data", path.Join(fileOutput, "/struct.go"), jsonFields)
	if err != nil {
		panic("json2Struct error:" + err.Error())
	}

	tpl, err := parseTpl(structDatas)
	if err != nil {
		panic("parseTpl" + err.Error())
	}

	err = writeFiles(tpl, fileOutput)
	if err != nil {
		panic("writeFiles" + err.Error())
	}

	fmt.Printf("[success] files has been saved in the %s directory\n", fileOutput)
}

func convertCfd(ckdMap map[string]conf.KeyData) map[string]model.ConfData {
	cfdMap := map[string]model.ConfData{}
	var cfd model.ConfData
	for k, ckd := range ckdMap {
		cfd = ckd.ConfData
		cfdMap[k] = cfd
	}
	return cfdMap
}

func getLibraries(sd model.StructData) []string {
	var libraries []string
	existLibraries := map[string]bool{}
	for _, cd := range sd.Data {
		library, ok := libraryMap[cd.Type]
		if ok {
			if _, exist := existLibraries[library]; !exist {
				libraries = append(libraries, library)
				existLibraries[cd.Type] = true
			}
		}
	}
	return libraries
}

func renderError(err error) {
	fmt.Println("usage: ")
	fmt.Println("from network: ./gokeeper-cli -d [domain name] -o [output directory] -k [keeper address] -s n")
	fmt.Println("from local: ./gokeeper-cli -i [input directory] -o [output directory] -s [l]")
	fmt.Println(err)
	os.Exit(1)
}
