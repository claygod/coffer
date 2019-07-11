package filenamer

// Coffer
// Filenamer
// Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
	"sync"
	//"time"
)

//const limitRecordsPerLogfile int64 = 100000

type FileNamer struct {
	m       sync.Mutex
	dirPath string
}

func NewFileNamer(dirPath string) *FileNamer {
	return &FileNamer{
		dirPath: dirPath,
	}
}

// func (f *FileNamer) GetNewFileName222(ext string) (string, error) {
// 	for i := 0; i < 60; i++ {
// 		if latestName, err := f.findLatestLog(ext); err == nil {
// 			lNumStr := strings.Replace(latestName, ext, "", 1)
// 			if lNum, err := strconv.ParseInt(lNumStr, 10, 64); err == nil {
// 				lNum += 1
// 				return f.dirPath + strconv.FormatInt(lNum, 10) + ext, nil // "/" +
// 			}
// 		}

// 		// newFileName := f.dirPath + strconv.Itoa(int(time.Now().Unix())) + ext
// 		// if _, err := os.Stat(newFileName); !os.IsExist(err) {
// 		// 	return newFileName, nil
// 		// }
// 		time.Sleep(1 * time.Second)
// 	}
// 	return "", fmt.Errorf("Error finding a new name.")
// }

func (f *FileNamer) GetNewFileName(ext string) (string, error) {
	f.m.Lock()
	defer f.m.Unlock()
	latestNum, err := f.findLatestNum()
	if err != nil {
		return "", fmt.Errorf("Error finding a new name: %v", err)
	}
	return f.dirPath + strconv.FormatInt(latestNum+1, 10) + ext, nil
}

func (f *FileNamer) GetLatestFileName(ext string) (string, error) {
	f.m.Lock()
	defer f.m.Unlock()
	//TODO:
	fNamesList, err := f.getFilesByExtList(ext)
	if err != nil {
		return "", fmt.Errorf("Error finding a latest name: %v", err)
	}
	ln := len(fNamesList)
	switch {
	case ln == 0:
		return "", nil
	case ln == 1:
		return fNamesList[0], nil
	default:
		sort.Strings(fNamesList)
		return fNamesList[len(fNamesList)-1], nil
	}
}

func (f *FileNamer) findLatestNum() (int64, error) {
	var max int64
	extList := []string{".log", ".check", ".checkout"} //TODO: нужен ли ".check" ???
	for _, ext := range extList {
		latestName, err := f.findLatestFile(ext)
		fmt.Println("LATEST: ", latestName)
		if err != nil {
			return 0, err
		} else if latestName == "" {
			continue
		}
		strs := strings.Split(latestName, ".")
		if len(strs) == 0 {
			continue
		}
		num, err := strconv.ParseInt(strs[0], 10, 64)
		if err == nil && num > max {
			max = num
		}
	}
	return max, nil
}

func (f *FileNamer) findLatestFile(ext string) (string, error) {
	fNamesList, err := f.getFilesByExtList(ext)
	if err != nil {
		return "", err
	}
	ln := len(fNamesList)
	switch {
	case ln == 0:
		return "", nil
	case ln == 1:
		return fNamesList[0], nil
	default:
		sort.Strings(fNamesList)
		return fNamesList[len(fNamesList)-1], nil
	}
	//return fNamesList, nil
}

// func (f *FileNamer) findLatestLog(ext string) (string, error) {
// 	fNamesList, err := f.getFilesByExtList(ext)
// 	if err != nil {
// 		return "", err
// 	}
// 	ln := len(fNamesList)
// 	switch {
// 	case ln == 0:
// 		return "0" + ext, nil
// 	case ln == 1: // последний лог мы никогда не берём чтобы не ткнуться в ещё наполняемый лог
// 		return fNamesList[0], nil
// 	default:
// 		sort.Strings(fNamesList)
// 		return fNamesList[len(fNamesList)-1], nil
// 	}
// 	//return fNamesList, nil
// }

func (f *FileNamer) getFilesByExtList(ext string) ([]string, error) {
	files, err := ioutil.ReadDir(f.dirPath)
	if err != nil {
		return nil, err
	}
	list := make([]string, 0, len(files))
	for _, fl := range files {
		if strings.HasSuffix(fl.Name(), ext) {
			list = append(list, fl.Name())
		}
	}
	return list, nil
}
