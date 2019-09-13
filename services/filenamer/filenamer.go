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
	//if ext == ".log" {
	latestNum, err := f.findLatestNum([]string{".log", ".check", ".checkpoint"})
	if err != nil {
		return "", fmt.Errorf("Error finding a new name: %v", err)
	}
	return f.dirPath + strconv.FormatInt(latestNum+1, 10) + ext, nil
	// }
	// latestNumLog, err := f.findLatestNum([]string{".log"}) // для ".checkpoint"
	// if err != nil {
	// 	return "", fmt.Errorf("Error finding a new name: %v", err)
	// }
	// latestNumChPn, err := f.findLatestNum([]string{".check", ".checkpoint"}) // для ".checkpoint"
	// if err != nil {
	// 	return "", fmt.Errorf("Error finding a new name: %v", err)
	// }
	// if latestNumLog > latestNumChPn {
	// 	return f.dirPath + strconv.FormatInt(latestNumLog, 10) + ext, nil
	// }
	// return f.dirPath + strconv.FormatInt(latestNumChPn+1, 10) + ext, nil

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
		return f.dirPath + fNamesList[0], nil
	default:
		sort.Strings(fNamesList)
		return f.dirPath + fNamesList[len(fNamesList)-1], nil
	}
}

func (f *FileNamer) GetAfterLatest(last string) ([]string, error) { //TODO: тут названия файлов возвращаются БЕЗ директории
	f.m.Lock()
	defer f.m.Unlock()
	lstTemp := strings.Split(last, "/") // на случай, если аргумент прилетел вместе с путём (директорией)
	lst := strings.Split(lstTemp[len(lstTemp)-1], ".")
	lastInt, err := strconv.Atoi(lst[0]) //      strconv.ParseInt(fNumStr, 10, 64)
	if err != nil || len(lst) != 2 {
		return nil, fmt.Errorf("Filenamer parse string (%s) error: %v", last, err)
	}

	fNamesList, err := f.getFilesByExtList(lst[1])
	if err != nil {
		return nil, fmt.Errorf("Error finding files by ext: %v", err)
	}
	//fmt.Println("FileNamer: fNamesList: ", lastInt, lst, fNamesList)
	numList := make([]int, 0, len(fNamesList))
	for _, fName := range fNamesList {
		fNumStr := strings.Replace(fName, "."+lst[1], "", 1)
		fNumInt, err := strconv.Atoi(fNumStr) //      strconv.ParseInt(fNumStr, 10, 64)
		if err != nil {
			//TODO: info
			continue
		}
		numList = append(numList, fNumInt)
	}
	sort.Ints(numList)
	//fmt.Println("FileNamer: numList: ", numList)
	outListInt := make([]int, 0, len(numList))
	for i, fNum := range numList {
		if fNum > lastInt {
			outListInt = numList[i : len(numList)-1]
			break
		}
	}
	//fmt.Println("FileNamer: outListInt: ", outListInt)
	outListStr := make([]string, 0, len(outListInt))
	for _, v := range outListInt {
		outListStr = append(outListStr, strconv.Itoa(v)+"."+lst[1])
	}
	//fmt.Println("FileNamer: outListStr: ", outListStr)
	return outListStr, nil
}

/*
GetHalf - получить список файлов, номера которых больше или меньше того, что в аргументе
*/
func (f *FileNamer) GetHalf(last string, more bool) ([]string, error) { //TODO: тут названия файлов возвращаются БЕЗ директории
	f.m.Lock()
	defer f.m.Unlock()
	lstTemp := strings.Split(last, "/") // на случай, если аргумент прилетел вместе с путём (директорией)
	lst := strings.Split(lstTemp[len(lstTemp)-1], ".")
	lastInt, err := strconv.Atoi(lst[0]) //      strconv.ParseInt(fNumStr, 10, 64)
	if err != nil || len(lst) != 2 {
		return nil, fmt.Errorf("Filenamer parse string (%s) error: %v", last, err)
	}

	fNamesList, err := f.getFilesByExtList(lst[1])
	if err != nil {
		return nil, fmt.Errorf("Error finding files by ext: %v", err)
	}
	//fmt.Println("FileNamer: fNamesList: ", lastInt, lst, fNamesList)
	numList := make([]int, 0, len(fNamesList))
	for _, fName := range fNamesList {
		fNumStr := strings.Replace(fName, "."+lst[1], "", 1)
		fNumInt, err := strconv.Atoi(fNumStr) //      strconv.ParseInt(fNumStr, 10, 64)
		if err != nil {
			//TODO: info
			continue
		}
		numList = append(numList, fNumInt)
	}
	sort.Ints(numList)
	//fmt.Println("FileNamer: ============ numList: ", numList)
	outListInt := make([]int, 0, len(numList))
	for i, fNum := range numList {
		if more && fNum > lastInt {
			outListInt = numList[i:len(numList)]
			break
		} else if !more && fNum >= lastInt {
			outListInt = numList[0:i]
			break
		}
	}
	//fmt.Println("FileNamer: outListInt: ", outListInt)
	outListStr := make([]string, 0, len(outListInt))
	for _, v := range outListInt {
		outListStr = append(outListStr, strconv.Itoa(v)+"."+lst[1])
	}
	//fmt.Println("FileNamer: outListStr: ", outListStr)
	return outListStr, nil
}

func (f *FileNamer) findLatestNum(extList []string) (int64, error) {
	var max int64
	//extList := []string{".log", ".check", ".checkpoint"} //TODO: нужен ли ".check" ???
	for _, ext := range extList {
		num, err := f.findMaxFile(ext)
		if err == nil && num > max {
			max = num
		}
		if err != nil {
			continue
		}

		//latestName, err := f.findLatestFile(ext)
		//fmt.Println("LATEST: ", num)

		// if err != nil {
		// 	return 0, err
		// } else if latestName == "" {
		// 	continue
		// }
		// strs := strings.Split(latestName, ".")
		// if len(strs) == 0 {
		// 	continue
		// }
		// num, err := strconv.ParseInt(strs[0], 10, 64)
		// if err == nil && num > max {
		// 	max = num
		// }
	}
	return max, nil
}

// func (f *FileNamer) findLatestFile(ext string) (string, error) {
// 	fNamesList, err := f.getFilesByExtList(ext)
// 	if err != nil {
// 		return "", err
// 	}
// 	ln := len(fNamesList)
// 	switch {
// 	case ln == 0:
// 		return "", nil
// 	case ln == 1:
// 		return fNamesList[0], nil
// 	default:
// 		sort.Strings(fNamesList)
// 		return fNamesList[len(fNamesList)-1], nil
// 	}
// 	//return fNamesList, nil
// }

func (f *FileNamer) findMaxFile(ext string) (int64, error) {
	fNamesList, err := f.getFilesByExtList(ext)
	if err != nil {
		return 0, err
	}
	ln := len(fNamesList)
	switch {
	case ln == 0:
		return 0, nil
	// case ln == 1:
	// 	return fNamesList[0], nil
	default:
		var max int64
		for _, name := range fNamesList {
			strs := strings.Split(name, ".")
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
