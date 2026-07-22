" rover.vim - Vim plugin for JDK source search, AST outline, symbol navigation, and markdown tools
if exists("g:roverpluginloaded")
  finish
endif
let g:roverpluginloaded=1

let s:plugindir = expand(expand("<sfile>:p:h:h"))
let s:binary = s:plugindir . "/.bin/rover"

" Commands for binary installation and updates
command! -nargs=* -complete=customlist,s:complete RoverInstall call s:RoverInstallBinaries(-1, <f-args>)
command! -nargs=* -complete=customlist,s:complete RoverUpdate  call s:RoverInstallBinaries(1, <f-args>)

function! s:RoverInstallBinaries(updateBinaries, ...)
  silent !clear
  execute "silent !" . "cd " . s:plugindir . "; " . "go build -o" . " " . s:binary
  echomsg "updated rover.vim plugin binary"
endfunction

" Check binary existence
function! s:ensureBinary()
  if !executable(s:binary)
    call s:RoverInstallBinaries(1)
  endif
  return executable(s:binary)
endfunction

" Document AST Outline (:RoverOutline / :JdkOutline)
function! s:RoverOutline()
  if !s:ensureBinary()
    return
  endif

  let l:file = expand('%:p')
  if empty(l:file)
    echoerr "Please save the file first"
    return
  endif

  let l:cmd = shellescape(s:binary) . " outline " . shellescape(l:file)
  let l:res = system(l:cmd)
  if v:shell_error != 0
    echoerr l:res
    return
  endif

  let l:list = json_decode(l:res)
  if empty(l:list)
    echomsg "No outline symbols found in " . expand('%:t')
    return
  endif

  call setloclist(0, l:list, 'r')
  lopen
endfunction

command! -nargs=0 RoverOutline call s:RoverOutline()
command! -nargs=0 JdkOutline   call s:RoverOutline()

" JDK Source Search (:JdkSearch / :RoverJdkSearch)
function! s:JdkSearch(...)
  if !s:ensureBinary()
    return
  endif

  let l:query = join(a:000, ' ')
  if empty(l:query)
    echoerr "Usage: :JdkSearch <query>"
    return
  endif

  let l:cmd = shellescape(s:binary) . " jdk-search " . shellescape(l:query)
  let l:res = system(l:cmd)
  if v:shell_error != 0
    echoerr l:res
    return
  endif

  let l:list = json_decode(l:res)
  if empty(l:list)
    echomsg "No JDK source matches found for: " . l:query
    return
  endif

  call setqflist(l:list, 'r')
  copen
endfunction

command! -nargs=* JdkSearch      call s:JdkSearch(<f-args>)
command! -nargs=* RoverJdkSearch call s:JdkSearch(<f-args>)

" Goto Symbol Declaration (:RoverGoto [symbol])
function! s:RoverGoto(...)
  if !s:ensureBinary()
    return
  endif

  let l:symbol = a:0 > 0 ? join(a:000, ' ') : expand('<cword>')
  if empty(l:symbol)
    echoerr "Usage: :RoverGoto <symbol>"
    return
  endif

  let l:cmd = shellescape(s:binary) . " goto " . shellescape(l:symbol)
  let l:res = system(l:cmd)
  if v:shell_error != 0
    echoerr l:res
    return
  endif

  let l:list = json_decode(l:res)
  if empty(l:list)
    echomsg "No declaration found for symbol: " . l:symbol
    return
  endif

  if len(l:list) == 1
    execute "edit +" . l:list[0].lnum . " " . fnameescape(l:list[0].filename)
  else
    call setqflist(l:list, 'r')
    copen
  endif
endfunction

command! -nargs=* RoverGoto call s:RoverGoto(<f-args>)

" JDK Cache Clean (:JdkClean / :RoverJdkClean)
function! s:JdkClean()
  if !s:ensureBinary()
    return
  endif

  let l:cmd = shellescape(s:binary) . " jdk-clean"
  let l:res = system(l:cmd)
  echomsg trim(l:res)
endfunction

command! -nargs=0 JdkClean      call s:JdkClean()
command! -nargs=0 RoverJdkClean call s:JdkClean()

" Markdown Image Paste (:MarkdownImagePaste)
function! s:MarkdownImagePaste()
  if !s:ensureBinary()
    return
  endif

  let l:file = expand('%:p')
  if empty(l:file)
    echoerr "Please save the file first"
    return
  endif

  let l:cmd = shellescape(s:binary) . " image-paste " . shellescape(l:file)
  let l:res = system(l:cmd)
  if v:shell_error != 0
    echoerr l:res
    return
  endif

  let l:tag = trim(l:res)
  if !empty(l:tag)
    execute "normal A" . l:tag
  endif
endfunction

command! -nargs=0 MarkdownImagePaste call s:MarkdownImagePaste()

" Markdown Preview (:MarkdownPreview)
function! s:MarkdownPreview()
  if !s:ensureBinary()
    return
  endif

  let l:file = expand('%:p')
  if empty(l:file)
    echoerr "Please save the file first"
    return
  endif

  let l:cmd = shellescape(s:binary) . " preview " . shellescape(l:file)
  call system(l:cmd)
endfunction

command! -nargs=0 MarkdownPreview call s:MarkdownPreview()

" Auto-clean JDK cache on Vim exit
function! s:AutoCleanJdkCache()
  if executable(s:binary)
    call system(shellescape(s:binary) . " jdk-clean")
  endif
endfunction

augroup RoverAutoClean
  autocmd!
  autocmd VimLeavePre * call s:AutoCleanJdkCache()
augroup END