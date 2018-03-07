package main

import (
	"os"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"time"
	"github.com/develar/zstd-archiver/tar"
)

func main() {
	app := kingpin.New("zstd-archiver", "Compress and decompress directory using tar and zstd").Version("0.0.1")

	ConfigureCompressCommand(app)
	ConfigureDecompressCommand(app)

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
}

func ConfigureCompressCommand(app *kingpin.Application) {
	command := app.Command("compress", "compress directory to tar.zst file").Alias("c")
	source := command.Flag("source", "source directory").Short('s').Required().String()
	output := command.Flag("output", "output file").Short('o').Required().String()
	level := command.Flag("level", "compression level (1-19)").Short('l').Default("19").Int()
	// 64MB produces the same size, but 16MB produces 200KB more, so, use 32 as default
	blockSize := command.Flag("block-size", "solid block size").Default("32MB").Bytes()
	isSystemIndependentFileInfo := command.Flag("system-independent", "whether to ignore time, xattr and other system dependent info").Bool()

	command.Action(func(context *kingpin.ParseContext) error {
		start := time.Now()
		err := tarutil.Compress(*source, *output, tarutil.CompressionOptions{
			Level: *level,
			BlockSize: int(*blockSize),
			IsSystemIndependentFileInfo: *isSystemIndependentFileInfo,
		})
		elapsed := time.Since(start)
		log.Printf("Compress took %s", elapsed)
		return err
	})
}

func ConfigureDecompressCommand(app *kingpin.Application) {
	command := app.Command("decompress", "decompress tar.zst file").Alias("d")
	source := command.Flag("input", "input file").Short('i').Required().String()
	output := command.Flag("output", "output directory").Short('o').Required().String()

	command.Action(func(context *kingpin.ParseContext) error {
		start := time.Now()
		err := tarutil.Decompress(*source, *output)
		elapsed := time.Since(start)
		log.Printf("Decompress took %s", elapsed)
		return err
	})
}