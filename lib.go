package randomfiles

import (
	"crypto/md5"
	"fmt"
	"hash"
	"io"
	"math/rand"
	"os"
	"path"
)

type Options struct {
	Out    io.Writer // output progress
	Source io.Reader // randomness source

	FileSize int    // the size per file.
	MinSize  int    // the size per file.
	Alphabet []rune // for filenames

	FanoutDepth int // how deep the hierarchy goes
	FanoutFiles int // how many files per dir
	FanoutDirs  int // how many dirs per dir

	RandomSeed   int64 // use a random seed. if 0, use a random seed
	RandomSize   bool  // randomize file sizes
	RandomFanout bool  // randomize fanout numbers

	MD5 bool //if defined, calculate MD5 for each generated file
}

func WriteRandomFiles(root string, depth int, opts *Options) error {

	numfiles := opts.FanoutFiles
	if opts.RandomFanout {
		numfiles = rand.Intn(numfiles) + 1
	}

	for i := 0; i < numfiles; i++ {
		if err := WriteRandomFile(root, opts); err != nil {
			return err
		}
	}

	if depth+1 <= opts.FanoutDepth {
		numdirs := opts.FanoutDirs
		if opts.RandomFanout {
			numdirs = rand.Intn(numdirs) + 1
		}

		for i := 0; i < numdirs; i++ {
			if err := WriteRandomDir(root, depth+1, opts); err != nil {
				return err
			}
		}
	}

	return nil
}

var FilenameSize = 16
var RunesEasy = []rune("abcdefghijklmnopqrstuvwxyz01234567890-_")
var RunesHard = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890!@#$%^&*()-_+= ;.,<>'\"[]{}() ")

func RandomFilename(length int, alphabet []rune) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(b)
}

func WriteRandomFile(root string, opts *Options) error {
	filesize := int64(opts.FileSize)
	if opts.RandomSize {
		minsz := int64(opts.MinSize)
		filesize = rand.Int63n(filesize-minsz) + minsz
	}

	n := rand.Intn(FilenameSize-4) + 4
	name := RandomFilename(n, opts.Alphabet)
	filepath := path.Join(root, name)
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	var dst io.Writer = f
	var hw hash.Hash
	if opts.MD5 {
		hw = md5.New()
		dst = io.MultiWriter(hw, f)
	}
	if _, err := io.CopyN(dst, opts.Source, filesize); err != nil {
		return err
	}

	if opts.MD5 {
		fmt.Fprintf(opts.Out, "%x\t", hw.Sum(nil))
	}
	if opts.Out != nil {
		fmt.Fprintln(opts.Out, filepath)
	}

	return f.Close()
}

func WriteRandomDir(root string, depth int, opts *Options) error {
	if depth > opts.FanoutDepth {
		return nil
	}

	n := rand.Intn(FilenameSize-4) + 4
	name := RandomFilename(n, opts.Alphabet)
	root = path.Join(root, name)
	if err := os.MkdirAll(root, 0755); err != nil {
		return err
	}

	if opts.Out != nil {
		fmt.Fprintln(opts.Out, root)
	}

	return WriteRandomFiles(root, depth, opts)
}
