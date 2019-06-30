package maillog_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/rojakcoder/mailgun/maillog"
	"github.com/stretchr/testify/assert"
)

func TestNewEmpty(t *testing.T) {

	l, err := maillog.New("", "").Init()
	assert.NotNil(t, l)
	assert.Nil(t, err)
	l.Errorf("hello %d", 4711)
	wc := l.NewWriter()
	n, err := wc.Write([]byte("H3ll0"))
	assert.NoError(t, err)
	assert.Exactly(t, 5, n)
}

// Tests runs as root on new caddy website ... not really sure if this test even
// makes sense.
func TestNewFail(t *testing.T) {
	t.Skip("TODO")
	testDir := path.Join(string(os.PathSeparator), "testdata") // try to create dir in root
	l, err := maillog.New(testDir, testDir).Init()
	assert.NotNil(t, l)
	assert.EqualError(t, err, "Cannot create directory \"/testdata\" because of: mkdir /testdata: permission denied")
}

func TestLogger_ErrDir_File(t *testing.T) {

	testDir := path.Join(".", "testdata", time.Now().String())
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Fatal(err)
		}
	}()
	l, err := maillog.New("", testDir).Init()
	if err != nil {
		t.Fatal(err)
	}

	const testData = `Snowden: The @FBI is creating a world where citizens rely on #Apple to defend their rights, rather than the other way around. https://t.co/vdjB6CuB7k`
	l.Errorf(testData)

	logContent, err := ioutil.ReadFile(l.ErrFile)
	if err != nil {
		t.Fatal(err)
	}
	assert.Contains(t, string(logContent), testData)
}

func TestLogger_MailDir_File(t *testing.T) {

	testDir := path.Join(".", "testdata", time.Now().String())
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Fatal(err)
		}
	}()
	l, err := maillog.New(testDir, "").Init("http://schumacherfm.local")
	if err != nil {
		t.Fatal(err)
	}

	wc := l.NewWriter()
	defer func() {
		if err := wc.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	var testData = []byte(`Snowden: The @FBI is creating a world where citizens rely on #Apple to defend their rights, rather than the other way around. https://t.co/vdjB6CuB7k`)
	n, err := wc.Write(testData)
	if err != nil {
		t.Fatal(err)
	}
	assert.Exactly(t, len(testData), n)
}

func TestLogger_MailDir_Stderr(t *testing.T) {
	orgStdErr := os.Stderr
	defer func() {
		os.Stderr = orgStdErr
	}()
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = pw
	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, pr); err != nil {
			t.Fatal(err)
		}
		outC <- buf.String()
	}()

	l, err := maillog.New("stderr", "").Init("http://schumacherfm.local")
	if err != nil {
		t.Fatal(err)
	}

	wc := l.NewWriter()
	defer wc.Close()

	var testData = []byte(`Snowden: The @FBI is creating a world where citizens rely on #Apple to defend their rights, rather than the other way around. https://t.co/vdjB6CuB7k`)
	n, err := wc.Write(testData)
	if err != nil {
		t.Fatal(err)
	}
	assert.Exactly(t, len(testData), n)

	// once finished with writing we must close the pipe
	if err := pw.Close(); err != nil {
		t.Fatal(err)
	}
	assert.Exactly(t, string(testData), <-outC)
}

func TestLogger_MailDir_Stdout(t *testing.T) {
	orgStdout := os.Stdout
	defer func() {
		os.Stdout = orgStdout
	}()
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = pw
	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, pr); err != nil {
			t.Fatal(err)
		}
		outC <- buf.String()
	}()

	l, err := maillog.New("stdout", "").Init("http://schumacherfm.local")
	if err != nil {
		t.Fatal(err)
	}

	wc := l.NewWriter()
	defer wc.Close()

	var testData = []byte(`Snowden: The @FBI is creating a world where citizens rely on #Apple to defend their rights, rather than the other way around. https://t.co/vdjB6CuB7k`)
	n, err := wc.Write(testData)
	if err != nil {
		t.Fatal(err)
	}
	assert.Exactly(t, len(testData), n)

	// once finished with writing we must close the pipe
	if err := pw.Close(); err != nil {
		t.Fatal(err)
	}
	assert.Exactly(t, string(testData), <-outC)
}

func TestLogger_ErrDir_Stderr(t *testing.T) {
	orgStdErr := os.Stderr
	defer func() {
		os.Stderr = orgStdErr
	}()
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = pw
	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, pr); err != nil {
			t.Fatal(err)
		}
		outC <- buf.String()
	}()

	l, err := maillog.New("", "stderr").Init()
	if err != nil {
		t.Fatal(err)
	}

	const testData = `Snowden: The @FBI is creating a world where citizens rely on #Apple to defend their rights, rather than the other way around. https://t.co/vdjB6CuB7k`
	l.Errorf(testData)

	// once finished with writing we must close the pipe
	if err := pw.Close(); err != nil {
		t.Fatal(err)
	}
	assert.Contains(t, <-outC, testData)

}
