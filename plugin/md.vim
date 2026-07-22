let s:plugindir = expand('<sfile>:p:h:h')

" These commands are available on any filetypes
command! -nargs=* -complete=customlist,s:complete RoverInstall call s:RoverInstallBinaries(-1, <f-args>)
command! -nargs=* -complete=customlist,s:complete RoverUpdate  call s:RoverInstallBinaries(1, <f-args>)
command! -nargs=* -complete=customlist,s:complete VimMdInstall call s:RoverInstallBinaries(-1, <f-args>)
command! -nargs=* -complete=customlist,s:complete VimMdUpdate  call s:RoverInstallBinaries(1, <f-args>)

function s:RoverInstallBinaries(updateBinaries, ...)
  let binary = ".bin/rover"

  silent !clear
  execute "silent !" . "cd " . s:plugindir . "; " . "go build -o" . " " . binary
  execute "silent !" . "cd " . s:plugindir . "; " . "cp -f " . binary . " .bin/vim-md"
  echomsg "updated rover.vim plugin"
endfunction