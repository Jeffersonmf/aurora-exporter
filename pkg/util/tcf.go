package util

// import "fmt"

type Block struct {
	Try     func()
	Catch   func(Exception)
	Finally func()
}

type Exception interface{}

func Throw(up Exception) {
	panic(up)
}

func (tcf Block) Do() {
	if tcf.Finally != nil {

		defer tcf.Finally()
	}
	if tcf.Catch != nil {
		defer func() {
			if r := recover(); r != nil {
				tcf.Catch(r)
			}
		}()
	}
	tcf.Try()
}

// func main() {
//
// 	fmt.Println("We Started")
//
// 	Block{
// 		Try: func() {
// 			fmt.Println("I tried")
// 			Throw("Ho, OMG....!!!")
// 		},
// 		Catch: func(e Exception) {
// 			fmt.Println("Caught %v\n", e)
// 		},
// 		Finally: func() {
// 			fmt.Println("Finally...")
// 		},
// 	}.Do()
//
// 	fmt.Println("We went on")
// }
//
