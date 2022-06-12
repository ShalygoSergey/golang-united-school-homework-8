package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

const permission = 0644
const Operation = "operation"
const OperationList = "list"
const OperationAdd = "add"
const OperationRemove = "remove"
const OperationFindById = "findById"
const FileName = "fileName"
const Item = "item"
const Id = "id"

type Arguments map[string]string

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func Perform(args Arguments, writer io.Writer) error {

	err := validate(args)
	if err != nil {
		return err
	}

	switch args[Operation] {
	case OperationList:
		return list(args[FileName], writer)
	case OperationAdd:
		return add(args[FileName], args[Item], writer)
	case OperationRemove:
		return remove(args[FileName], args[Id], writer)
	case OperationFindById:
		return find(args[FileName], args[Id], writer)
	}

	return nil
}

func validate(args Arguments) error {
	operation, ok := args[Operation]

	if !ok || operation == "" {
		return errors.New("-operation flag has to be specified")
	}

	if operation != OperationList && operation != OperationAdd && operation != OperationRemove && operation != OperationFindById {
		return fmt.Errorf("Operation %s not allowed!", operation)
	}

	fileName, ok := args[FileName]

	if !ok || fileName == "" {
		return errors.New("-fileName flag has to be specified")
	}

	switch operation {
	case OperationAdd:
		item, ok := args[Item]

		if !ok || item == "" {
			return errors.New("-item flag has to be specified")
		}
	case OperationRemove, OperationFindById:
		id, ok := args[Id]

		fmt.Println(id)
		fmt.Println(ok)

		if !ok || id == "" {
			return errors.New("-id flag has to be specified")
		}
	}

	return nil
}

func find(fileName string, id string, writer io.Writer) error {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, permission)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

	if err != nil {
		return err
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	var users []User
	if len(bytes) != 0 {
		err = json.Unmarshal(bytes, &users)
		if err != nil {
			return err
		}
	}

	for _, v := range users {
		if v.Id == id {
			jsonMarshal, err := json.Marshal(&v)
			if err != nil {
				return err
			}

			err = file.Truncate(0)
			if err != nil {
				return err
			}

			_, err = file.Seek(0, 0)
			if err != nil {
				return err
			}

			_, err = writer.Write(jsonMarshal)
			if err != nil {
				return err
			}

			return nil
		}
	}

	err = file.Truncate(0)
	if err != nil {
		return err
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = file.Write([]byte("\n"))
	if err != nil {
		return err
	}

	return nil
}

func remove(fileName string, id string, writer io.Writer) error {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, permission)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

	if err != nil {
		return err
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	var users []User
	if len(bytes) != 0 {
		err = json.Unmarshal(bytes, &users)
		if err != nil {
			return err
		}
	}

	for i := 0; i < len(users); i++ {
		if users[i].Id == id {

			users = append(users[:i], users[i+1:]...)

			jsonMarshal, err := json.Marshal(&users)
			if err != nil {
				return err
			}

			err = file.Truncate(0)
			if err != nil {
				return err
			}

			_, err = file.Seek(0, 0)
			if err != nil {
				return err
			}

			_, err = file.Write(jsonMarshal)
			if err != nil {
				return err
			}

			return nil
		}
	}

	_, err = writer.Write([]byte(fmt.Sprintf("Item with id %s not found", id)))
	if err != nil {
		return err
	}

	return nil
}

func add(fileName string, item string, writer io.Writer) error {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, permission)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

	if err != nil {
		return err
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	var user User

	if len(item) != 0 {
		err = json.Unmarshal([]byte(item), &user)
		if err != nil {
			return err
		}
	}

	var users []User

	if len(bytes) != 0 {
		err = json.Unmarshal(bytes, &users)
		if err != nil {
			return err
		}
	}

	for _, v := range users {
		if v.Id == user.Id {
			_, err := writer.Write([]byte(fmt.Sprintf("Item with id %s already exists", user.Id)))
			if err != nil {
				return err
			}
			return nil
		}
	}

	users = append(users, user)

	jsonMarshal, err := json.Marshal(&users)
	if err != nil {
		return err
	}

	err = file.Truncate(0)
	if err != nil {
		return err
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = file.Write(jsonMarshal)
	if err != nil {
		return err
	}

	return nil
}

func list(fileName string, writer io.Writer) error {
	file, err := os.OpenFile(fileName, os.O_RDONLY, permission)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(file)

	if err != nil {
		return err
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	_, err = writer.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}

func parseArgs() Arguments {
	var args = make(Arguments)

	args[Id] = *flag.String(Id, "", Id)
	args[Item] = *flag.String(Item, "", Item)
	args[Operation] = *flag.String(Operation, "", Operation)
	args[FileName] = *flag.String(FileName, "", FileName)

	return args
}
