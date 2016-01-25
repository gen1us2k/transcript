package transcript

import (
	"testing"
)


func TestTranscript(t *testing.T){
	LoadDict("")
	s := "Jessie Jane"
	transcripted := TransliterateRussian(s)
	if transcripted != "джеси джен " {
		t.Fatal("Wrong transcription")
	}
}