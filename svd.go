package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/bbalet/stopwords"
	"github.com/gonum/matrix/mat64"
	"github.com/reiver/go-porterstemmer"
	"github.com/fatlotus/gauss"
)

type clustersData struct {
	Clusters       map[string][]int    `json:"Clusters"`
	ClustersLabels map[string]string   `json:"ClusterLabels"`
	ClustersFiles  map[string][]string `json:"ClusterFiles"`
}

type MatrixE struct {
	Terms 	[][]float64 	`json:"Terms"`
	Docs 	[][]float64  	`json:"Docs"`
}

type phrase struct {
	word   string
	wordID int
	DocID  []int
	count  int
	tf     []int
	df     int
	idf    float64
	tfidf  []float64
	wfl    bool
}

type similarity struct {
	dname     string
	simlarity float64
}
type Bysimilarity []similarity

var dID, wID int
var svd mat64.SVD

func main() {

	searchDir := "./extraction"

	firstTime := true //Will let the next loop know if this is the first clustering or not

	var docsToInclude []int //Will be the array with indices of documents to include in clustering

	var initialFileNames []string

	totalCount := 0
	var initialClusterArr [][]int

	var finalClusterArr [][]int
	var finalClusterArrLabels []string
	for totalCount < 1 {
		totalCount = totalCount + 1
		dID = -1
		wID = -1
		termdocumentMatrix := []float64{}
		var tosort map[string]int
		var sorted []string
		tosort = make(map[string]int)

		fileNames := []string{}
		var phrases map[string]phrase
		phrases = make(map[string]phrase)

		var docID map[int]string
		docID = make(map[int]string)

		if firstTime == false {
			docsToInclude = initialClusterArr[totalCount-2]
		}

		count := 0
		err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
			if strings.Contains(path, ".txt") {
				if firstTime == true {
					docsToInclude = append(docsToInclude, count)
				}
				if checker(docsToInclude, count) {
					fileNames = append(fileNames, path)
				}
				count = count + 1
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
		numeberofdocs := len(fileNames)

		if firstTime == true {
			initialFileNames = fileNames
		}

		for _, fileName := range fileNames[:] {
			//------------formating file names-----------
			r := strings.NewReplacer("+", " ",
				".txt", "")
			result := r.Replace(fileName)
			var re = regexp.MustCompile(`testdata2.+\/`)
			fname := re.ReplaceAllString(result, "")
			//------------file name formatting ends------------
			dID = dID + 1
			docID[dID] = fname
			frequency(fileName, phrases)
			for k := range phrases {
				tmp := phrases[k]
				if tmp.count > 0 {
					tmp.df++
					tmp.tf = append(tmp.tf, tmp.count)
					tmp.DocID = append(tmp.DocID, dID)
				}
				tmp.count = 0
				phrases[k] = tmp
			}
		}
		for k := range phrases {
			tmp := phrases[k]
			tmp.idf = math.Log10(float64(numeberofdocs) / float64(1+tmp.df))
			for _, v := range tmp.tf {
				tmp.tfidf = append(tmp.tfidf, tmp.idf*float64(v))
			}
			phrases[k] = tmp
			tosort[k] = tmp.wordID
		}
		sorted = sortmapbyvalue(tosort)
		for _, w := range sorted {
			termFreqs := make([]float64, numeberofdocs)
			tmp := phrases[w]
			for i, v := range tmp.DocID {
				termFreqs[v] = float64(toFixed(tmp.tfidf[i], 4))
			}
			termdocumentMatrix = append(termdocumentMatrix, termFreqs...)
		}

		//--------------------printing tfidf matrix---------------//
		mapOfMaps := make(map[int][]string, numeberofdocs)

		orderedPhrases := make([]string, len(phrases))
		counter := 0
		for k, _ := range phrases {
			orderedPhrases[counter] = k
			counter++
		}

		for i := 0; i < numeberofdocs; i++ {
			var m []string
			mapOfMaps[i] = m
		}
		for _, v := range phrases {
			for i := 0; i < len(v.DocID); i++ {
				var buffer bytes.Buffer
				buffer.WriteString(v.word)
				buffer.WriteString(":")
				buffer.WriteString(strconv.FormatFloat(v.tfidf[i], 'f', -1, 64))
				if v.tfidf[i] > .00000001 || v.tfidf[i] < -.00000001 {
					mapOfMaps[v.DocID[i]] = append(mapOfMaps[v.DocID[i]], buffer.String())
				}
			}

		}
		//---------------------printing done-----------------------//
		// fmt.Println(termdocumentMatrix)
		newtd := make([][]float64, len(phrases))
		for i := 0; i < len(phrases); i++{
			newtd[i] = make([]float64, numeberofdocs)
		}
		count = 0
		for i := 0; i < len(phrases); i++{
			for j := 0; j < numeberofdocs; j++{
				newtd[i][j] = termdocumentMatrix[count]
				count = count + 1
			}
		}
		fmt.Println("The SVD will take a few minutes...")
		uNew,sNew,vNew := gauss.SVD(gauss.Matrix(newtd))
		fmt.Println("sNew", sNew)

		fmt.Println(len(phrases))
		tdm2 := mat64.NewDense(len(phrases), numeberofdocs, termdocumentMatrix)
		fmt.Println(len(phrases))
		r1, c1 := tdm2.Dims()
		fmt.Println("TDM2: ", r1, c1)
		
		r := uNew.Shape[0]
		c := uNew.Shape[1]
		LRF := 330 // low rank factor 'sometimes refered to as 'k' in different texts'
		slen := len(sNew.Data) - LRF
		//-------------- extracting a low rank matrix ------------//
		TermE := mat64.NewDense(r, slen, nil)
		eigenvalues := mat64.NewDense(slen, slen, nil)
		c = vNew.Shape[1]
		DocE := mat64.NewDense(slen, c, nil)
		for i := 0; i < slen; i++ {
			eigenvalues.Set(i, i, sNew.Data[i])
		}
		vSmall := mat64.NewDense(slen, c, nil)
		r, c = vSmall.Dims()
		//fmt.Println(r, c)
		for i := 0; i < slen; i++ {
			for j := 0; j < c; j++ {
				vSmall.Set(i, j, vNew.Data[(j * slen) + i])
			}
		}
		r = uNew.Shape[0]
		c = uNew.Shape[1]
		uSmall := mat64.NewDense(r, slen, nil)
		count = 0
		for i := 0; i < r; i++ {
			for j := 0; j < slen; j++ {
				uSmall.Set(i, j, uNew.Data[count])
				count = count + 1
			}
		}
		//--------------------low rank extraction done -------------//
		//---------------now scaling the term space and doc space by the eigen values----------------//
		var matrices MatrixE
		TermE.Mul(uSmall, eigenvalues)
		DocE.Mul(eigenvalues, vSmall)
		
		Tr, Tc := TermE.Dims()
		Dr, Dc := DocE.Dims()

		matrices.Terms = make([][]float64, Tr)
		for i := range matrices.Terms {
			matrices.Terms[i] = make([]float64, Tc)
		}
		for i := 0; i < Tr; i++ {
			for j := 0; j < Tc; j++ {
				matrices.Terms[i][j] = TermE.At(i, j)
			}
		}
		matrices.Docs = make([][]float64, Dr)
		for i := range matrices.Docs {
			matrices.Docs[i] = make([]float64, Dc)
		}
		for i := 0; i < Dr; i++ {
			for j := 0; j < Dc; j++ {
				matrices.Docs[i][j] = DocE.At(i, j)
			}
		}

		//===========================================================
		if len(os.Args) > 1{
			printMatrix(DocE, mapOfMaps, fileNames, os.Args[1])
		} else {
			printMatrix(DocE, mapOfMaps, fileNames, "outputMatrix.txt")
		}
		//===========================================================
	}
	for i := 0; i < len(finalClusterArr); i++ {
		fmt.Println("--------------------------------------------------------")
		fmt.Println(finalClusterArr[i])
		for j := 0; j < len(finalClusterArr[i]); j++ {
			fileString := initialFileNames[finalClusterArr[i][j]]
			fileStringArr := strings.Split(fileString, "/")
			fmt.Println(fileStringArr[len(fileStringArr)-1])

			// fmt.Println(initialFileNames[finalClusterArr[i][j]])
		}
		fmt.Println(finalClusterArrLabels[i])
	}
}

//Prints out matrix in correct format for later python scripts
func printMatrix(matr *mat64.Dense, m map[int][]string, fileNames []string, outputFile string) {
	// f, _ := os.Create("outputMatrix.txt")
	f, _ := os.Create(outputFile)
	a, b := matr.Dims()
	Cols := make([][]float64, b)
	for i := 0; i < b; i++ {
		Cols[i] = make([]float64, a)
		CurrCol := matr.ColView(i)
		for j := 0; j < a; j++ {
			Cols[i][j] = CurrCol.At(j, 0)
			if Cols[i][j] < 0.0000001 && Cols[i][j] > 0 {
				Cols[i][j] = 0.0
			}
			if Cols[i][j] > -0.0000001 && Cols[i][j] < 0 {
				Cols[i][j] = 0.0
			}
		}
	}
	// fmt.Println("Cols:", Cols)
	for i := 0; i < b; i++ {
		for j := 0; j < a; j++ {
			// fmt.Print(Cols[i][j])
			f.WriteString(strconv.FormatFloat(Cols[i][j], 'f', -1, 64))
			if j < a-1 {
				// fmt.Print(",")
				f.WriteString(",")
			}
		}
		// fmt.Print(",")
		f.WriteString(",")
		for j := 0; j < len(m[i]); j++ {
			// fmt.Print(m[i][j],"|")
			f.WriteString(m[i][j])
			f.WriteString("|")
		}
		// fmt.Print(",", fileNames[i])
		// fmt.Print("\n")
		f.WriteString(",")
		f.WriteString(fileNames[i])
		f.WriteString("\n")
	}
}

//Check if element exists in array
func checker(arr []int, element int) bool {
	for i := 0; i < len(arr); i++ {
		if arr[i] == element {
			return true
		}
	}
	return false
}

//------------comment the following if output/directory is not required-------------//

//-------------------------clearing done ---------------------//

func frequency(fileName string, p map[string]phrase) {
	fileContent, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}
	fileL := strings.ToLower(string(fileContent))
	cleanContent := stopwords.CleanString(fileL, "en", true)
	s := strings.Fields(cleanContent)
	
	for _, i := range s {
		stem := porterstemmer.StemString(i)
		tmp := p[stem]
		tmp.word = stem
		tmp.count++
		if tmp.wfl != true {
			wID = wID + 1
			tmp.wordID = wID
			tmp.wfl = true
		}
		p[stem] = tmp
	}
}

//---------------------write output---------------------//
func writeoutput(sorted []string, phrases map[string]phrase) {
	fmt.Println("word wID dID count tf df idf tfidf wfl")
	for _, v := range sorted {
		fmt.Println(phrases[v])
	}

}

func del(m map[string]phrase) {
	for k := range m {
		delete(m, k)
	}
}
func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

//------------------------sorting starts here---------------------//
type sortedMap struct {
	m map[string]int
	s []string
}

func (sm *sortedMap) Len() int {
	return len(sm.m)
}

func (sm *sortedMap) Less(i, j int) bool {
	return sm.m[sm.s[i]] < sm.m[sm.s[j]]
}

func (sm *sortedMap) Swap(i, j int) {
	sm.s[i], sm.s[j] = sm.s[j], sm.s[i]
}

func sortmapbyvalue(m map[string]int) []string {
	sm := new(sortedMap)
	sm.m = m
	sm.s = make([]string, len(m))
	i := 0
	for key := range m {
		sm.s[i] = key
		i++
	}
	sort.Sort(sm)
	return sm.s
}

func (a Bysimilarity) Len() int           { return len(a) }
func (a Bysimilarity) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Bysimilarity) Less(i, j int) bool { return a[i].simlarity > a[j].simlarity }

//----------------------------------------------------------------------//

func printdense(m *mat64.Dense, name string) {
	r, c := m.Dims()
	fmt.Println("Matrix ", name, r, c)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			fmt.Printf("%f	", m.At(i, j))
		}
		fmt.Printf("\n")
	}
}

func cosinesimilarity(a *mat64.Vector, b *mat64.Vector, docname string) similarity {
	var sum float64
	sum = 0
	moda := 0.0
	modb := 0.0
	var tmp similarity
	for i := 0; i < a.Len(); i++ {
		moda += math.Pow(a.At(i, 0), 2)
		modb += math.Pow(b.At(i, 0), 2)
		sum += a.At(i, 0) * b.At(i, 0)
	}
	tmp.dname = docname
	tmp.simlarity = sum / (math.Sqrt(moda) * math.Sqrt(modb))
	// fmt.Println(tmp.simlarity)
	return tmp
}
