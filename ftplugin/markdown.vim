" Workaround for https://github.com/vim/vim/issues/4530
if exists("g:vimmdpluginloaded")
  finish
endif
let g:vimmdpluginloaded=1
let s:loadStatusCallbacks = []

let s:minVimSafeState = has("patch-8.1.2056")

augroup vimmd
augroup END

let s:channel = ""
"let s:timer = ""
let s:plugindir = expand(expand("<sfile>:p:h:h"))

let s:waitingToDrain = 0
let s:scheduleBacklog = []
let s:activeGovimCalls = 0

function s:define(channel, msg)
  " format is [id, type, ...]
  " type is function, command or autocmd
  try
    let l:id = a:msg[0]
    let l:resp = ["callback", l:id, [""]]
    if a:msg[1] == "loaded"
      let s:govim_status = "loaded"
      for F in s:loadStatusCallbacks
        call call(F, [s:govim_status])
      endfor
    elseif a:msg[1] == "initcomplete"
      let s:govim_status = "initcomplete"
      for F in s:loadStatusCallbacks
        call call(F, [s:govim_status])
      endfor
    elseif a:msg[1] == "currentViewport"
      let l:res = s:buildCurrentViewport()
    elseif a:msg[1] == "function"
      call s:defineFunction(a:msg[2], a:msg[3], 0)
    elseif a:msg[1] == "rangefunction"
      call s:defineFunction(a:msg[2], a:msg[3], 1)
    elseif a:msg[1] == "command"
      call s:defineCommand(a:msg[2], a:msg[3])
    elseif a:msg[1] == "autocmd"
      call s:defineAutoCommand(a:msg[2], a:msg[3], a:msg[4])
    elseif a:msg[1] == "redraw"
      let l:force = a:msg[2]
      let l:args = ""
      if l:force == "force"
        let l:args = '!'
      endif
      execute "redraw".l:args
    elseif a:msg[1] == "ex"
      let l:expr = a:msg[2]
      execute l:expr
    elseif a:msg[1] == "normal"
      let l:expr = a:msg[2]
      execute "normal ".l:expr
    elseif a:msg[1] == "expr"
      let l:expr = a:msg[2]
      let l:res = eval(l:expr)
      call add(l:resp[2], l:res)
    elseif a:msg[1] == "call"
      let l:fn = a:msg[2]
      let F= function(l:fn, a:msg[3:-1])
      let l:res = F()
      call add(l:resp[2], l:res)
    elseif a:msg[1] == "error"
      let l:msg = a:msg[2]
      " this is an async call from the client
      throw l:msg
      return
    else
      throw "unknown callback function type ".a:msg[1]
    endif
  catch
    let l:resp[2][0] = 'Caught ' . string(v:exception) . ' in ' . v:throwpoint
  endtry
  call ch_sendexpr(a:channel, l:resp)
endfunction

function s:ch_evalexpr(args)
  " For all callbacks to govim (other than the handler ultimately responsible
  " for a listener_add callback) we need to flush any pending delta
  " notifications so that govim isn't ever working with stale buffer
  " contents
  if a:args[0] != "function" || a:args[1] != "function:GOVIM_internal_BufChanged"
    call listener_flush()
  endif
  if s:minVimSafeState
    let l:resp = ch_evalexpr(s:channel, a:args)
    if l:resp[0] != ""
      throw l:resp[0]
    endif
    return l:resp[1]
  endif

  let s:activeGovimCalls += 1
  let l:resp = ch_evalexpr(s:channel, a:args)
  let s:activeGovimCalls -= 1
  if l:resp[0] != ""
    throw l:resp[0]
  endif
  call s:drainScheduleBacklog(v:false)
  return l:resp[1]
endfunction

function s:callbackCommand(name, flags, ...)
  let l:args = ["function", "command:".a:name, a:flags]
  call extend(l:args, a:000)
  return s:ch_evalexpr(l:args)
endfunction

func s:defineCommand(name, attrs)
  let l:def = "command! "
  let l:args = ""
  let l:flags = ['"mods": expand("<mods>")']
  " let l:flags = []
  if has_key(a:attrs, "nargs")
    let l:def .= " ". a:attrs["nargs"]
    if a:attrs["nargs"] != "-nargs=0"
      let l:args = ", <f-args>"
    endif
  endif
  if has_key(a:attrs, "range")
    let l:def .= " ".a:attrs["range"]
    call add(l:flags, '"line1": <line1>')
    call add(l:flags, '"line2": <line2>')
    call add(l:flags, '"range": <range>')
  endif
  if has_key(a:attrs, "count")
    let l:def .= " ". a:attrs["count"]
    call add(l:flags, '"count": <count>')
  endif
  if has_key(a:attrs, "complete")
    let l:def .= " ". a:attrs["complete"]
  endif
  if has_key(a:attrs, "general")
    for l:a in a:attrs["general"]
      let l:def .= " ". l:a
      if l:a == "-bang"
        call add(l:flags, '"bang": "<bang>"')
      endif
      if l:a == "-register"
        call add(l:flags, '"register": "<reg>"')
      endif
    endfor
  endif
  let l:flagsStr = "{" . join(l:flags, ", ") . "}"
  let l:def .= " " . a:name . " call s:callbackCommand(\"". a:name . "\", " . l:flagsStr . l:args . ")"
  execute l:def
endfunction

function s:doShutdown() 
  call ch_close(s:channel)
endfunction

let opts = {"in_mode": "json", "out_mode": "json", "err_mode": "json", "callback": function("s:define"), "timeout": 30000, "waittime": 5000} 
let targetdir = s:plugindir . "/.bin/"
let start = filereadable(targetdir . "rover") ? targetdir . "rover" : targetdir . "vim-md"

let job = job_start(start)
"let s:channel = job_getchannel(job)
let s:channel = ch_open("localhost:8765", opts)

" TODO remove me for debug purpose only
" echo "job started".ch_status(s:channel)

au VimLeave * call s:doShutdown()
