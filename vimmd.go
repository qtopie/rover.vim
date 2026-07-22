package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/govim/govim"
	"gopkg.in/tomb.v2"
)

type VimMdPlugin struct {
	tomb tomb.Tomb
}

func (p *VimMdPlugin) Init(g govim.Govim, ch chan error) error {
	g.DefineCommand("MarkdownPreview", previewMarkdown)
	g.DefineCommand("MarkdownImagePaste", pasteImage)
	g.DefineCommand("MarkdownImageClean", cleanImage)
	g.DefineCommand("JdkSearch", searchJdkSource, govim.NArgsZeroOrMore)
	g.DefineCommand("RoverJdkSearch", searchJdkSource, govim.NArgsZeroOrMore)
	g.DefineCommand("JdkClean", cleanJdkCache)
	g.DefineCommand("RoverJdkClean", cleanJdkCache)

	// Clean up JDK cache when Vim leaves/exits
	_ = g.DefineAutoCommand("RoverJdkAutoClean", govim.Events{govim.EventVimLeavePre, govim.EventVimLeave}, govim.Patterns{"*"}, false, func(g govim.Govim, args ...json.RawMessage) error {
		return removeJdkCacheDir()
	})

	// Check and clean JDK cache when closing/unloading buffers
	_ = g.DefineAutoCommand("RoverJdkBufAutoClean", govim.Events{govim.EventBufUnload, govim.EventBufWipeout}, govim.Patterns{"*.java"}, false, func(g govim.Govim, args ...json.RawMessage) error {
		return checkAndAutoCleanJdkCache(g)
	})

	return nil
}

func (p *VimMdPlugin) Shutdown() error {
	_ = removeJdkCacheDir()
	return nil
}

func showMsg(g govim.Govim, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return g.ChannelEx("echomsg " + strconv.Quote(msg))
}

func showErrMsg(g govim.Govim, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return g.ChannelEx("echoerr " + strconv.Quote(msg))
}

func appendLine(g govim.Govim, format string, args ...interface{}) error {
	line := fmt.Sprintf(format, args...)
	return g.ChannelNormal("A" + line)
}
