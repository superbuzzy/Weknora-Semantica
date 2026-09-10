package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)
func ReadJSON(path string,out interface{})error{data,err:=os.ReadFile(path);if err!=nil{return fmt.Errorf("read %s: %w",path,err)};if err:=json.Unmarshal(data,out);err!=nil{return fmt.Errorf("decode %s: %w",path,err)};return nil}
func WriteJSON(path string,value interface{})error{data,err:=json.MarshalIndent(value,"","  ");if err!=nil{return err};if err:=os.MkdirAll(filepath.Dir(path),0o755);err!=nil{return err};if err:=os.WriteFile(path,append(data,'\n'),0o644);err!=nil{return fmt.Errorf("write %s: %w",path,err)};return nil}
