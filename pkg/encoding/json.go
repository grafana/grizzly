package encoding

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func MarshalToJSON(spec map[string]interface{}, filename string) error {
	j, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(filename, j)
}

func writeFile(filename string, content []byte) error {
	dir := filepath.Dir(filename)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, content, 0644)
	if err != nil {
		return err
	}
	return nil
}
