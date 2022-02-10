---
title: Go 程序启动随笔
date: 2022-02-07 22:55:34
tags: 
- source
- go
categories:
- Golang
---

### 启动代码

```assembly
/*** go 1.17.6 ***/

/**************** [asm_amd64.s] ****************/

/* entry point */
TEXT _rt0_amd64(SB),NOSPLIT,$-8
	MOVQ	0(SP), DI	// argc
	LEAQ	8(SP), SI	// argv
	JMP	runtime·rt0_go(SB)
	
/* 主启动流程
 * 1. 该函数代表 runtime 包下的 rt0_go 函数，“·” 符号用于路径分隔 
 * 2. NOSPLIT = 不需要栈分割，TOPFRAME = 调用栈最顶层，Traceback 会在此停止
*/
TEXT runtime·rt0_go(SB),NOSPLIT|TOPFRAME,$0
	/* AX = argc， BX = argv */
	MOVQ	DI, AX		// argc
	MOVQ	SI, BX		// argv
	
	/* 扩张当前栈空间至 SP - (4*8+7)，再将 SP 地址按 16 字节对齐（部分 CPU 指令要求对齐，如 SSE）*/
	SUBQ	$(4*8+7), SP		// 2args 2auto
	ANDQ	$~15, SP
	
	/* SP+16 = argc， SP+24 = argv */
	MOVQ	AX, 16(SP)
	MOVQ	BX, 24(SP)

	/* 初始化 g0 的 stack，SB 伪寄存器配合前缀可得到 g0 在 DATA 区的地址 */
	MOVQ	$runtime·g0(SB), DI
	LEAQ	(-64*1024+104)(SP), BX
	MOVQ	BX, g_stackguard0(DI)
	MOVQ	BX, g_stackguard1(DI)
	/* g0 栈空间下限 = SP - 64Kib + 104byte，栈空间上限 = SP */
	MOVQ	BX, (g_stack+stack_lo)(DI)
	MOVQ	SP, (g_stack+stack_hi)(DI)

	/* CPU 信息设置以及 cgo 对 g0 栈空间的影响 */
	MOVL	$0, AX
	CPUID
	MOVL	AX, SI
	CMPL	AX, $0
	JE	nocpuinfo

	// Figure out how to serialize RDTSC.
	// On Intel processors LFENCE is enough. AMD requires MFENCE.
	// Don't know about the rest, so let's do MFENCE.
	CMPL	BX, $0x756E6547  // "Genu"
	JNE	notintel
	CMPL	DX, $0x49656E69  // "ineI"
	JNE	notintel
	CMPL	CX, $0x6C65746E  // "ntel"
	JNE	notintel
	MOVB	$1, runtime·isIntel(SB)
	MOVB	$1, runtime·lfenceBeforeRdtsc(SB)
notintel:

	// Load EAX=1 cpuid flags
	MOVL	$1, AX
	CPUID
	MOVL	AX, runtime·processorVersionInfo(SB)

nocpuinfo:
	// if there is an _cgo_init, call it.
	MOVQ	_cgo_init(SB), AX
	TESTQ	AX, AX
	JZ	needtls
	// arg 1: g0, already in DI
	MOVQ	$setg_gcc<>(SB), SI // arg 2: setg_gcc
#ifdef GOOS_android
	MOVQ	$runtime·tls_g(SB), DX 	// arg 3: &tls_g
	// arg 4: TLS base, stored in slot 0 (Android's TLS_SLOT_SELF).
	// Compensate for tls_g (+16).
	MOVQ	-16(TLS), CX
#else
	MOVQ	$0, DX	// arg 3, 4: not used when using platform's TLS
	MOVQ	$0, CX
#endif
#ifdef GOOS_windows
	// Adjust for the Win64 calling convention.
	MOVQ	CX, R9 // arg 4
	MOVQ	DX, R8 // arg 3
	MOVQ	SI, DX // arg 2
	MOVQ	DI, CX // arg 1
#endif
	CALL	AX

	// update stackguard after _cgo_init
	MOVQ	$runtime·g0(SB), CX
	MOVQ	(g_stack+stack_lo)(CX), AX
	ADDQ	$const__StackGuard, AX
	MOVQ	AX, g_stackguard0(CX)
	MOVQ	AX, g_stackguard1(CX)

/* 设置 TLS，部分 OS 直接跳过 */
#ifndef GOOS_windows
	JMP ok
#endif
needtls:
#ifdef GOOS_plan9
	// skip TLS setup on Plan 9
	JMP ok
#endif
#ifdef GOOS_solaris
	// skip TLS setup on Solaris
	JMP ok
#endif
#ifdef GOOS_illumos
	// skip TLS setup on illumos
	JMP ok
#endif
#ifdef GOOS_darwin
	// skip TLS setup on Darwin
	JMP ok
#endif
#ifdef GOOS_openbsd
	// skip TLS setup on OpenBSD
	JMP ok
#endif

  /* DI = m0 的 m_tls 字段 DATA 地址 */
	LEAQ	runtime·m0+m_tls(SB), DI
	/* settls 函数在 sys_linux_amd64.s 内
	 * 主要通过 arch_prctl 系统调用，将 m_tls 的地址设置到 FS 寄存器内
	*/
	CALL	runtime·settls(SB)

	/* 检查 TLS 是否成功设置：
   * get_tls(BX) 将当前 TLS 地址放入 BX （实际上是一个宏定义： #define	get_tls(r)	MOVQ TLS, r ）
   * 将 0x123 立即数存入 TLS，再从 m_tls 地址读出，如果相等说明立即数已经正确存入
  */
	get_tls(BX)
	MOVQ	$0x123, g(BX)
	MOVQ	runtime·m0+m_tls(SB), AX
	CMPQ	AX, $0x123
	JEQ 2(PC)
	CALL	runtime·abort(SB)
	
ok:
	/* 绑定 m0 和 g0 */
	get_tls(BX)
	LEAQ	runtime·g0(SB), CX
	MOVQ	CX, g(BX)
	LEAQ	runtime·m0(SB), AX

	// save m->g0 = g0
	MOVQ	CX, m_g0(AX)
	// save m0 to g0->m
	MOVQ	AX, g_m(CX)

	CLD				// convention is D is always left cleared
	/* 类型检查，见 runtime1.go: check() */
	CALL	runtime·check(SB)

  /* SP = argc, SP + 8 = argv, SP 和 SP + 8 作为调用下层函数 args 的输入参数（函数参数可见 FP 伪寄存器） */
	MOVL	16(SP), AX		// copy argc
	MOVL	AX, 0(SP)
	MOVQ	24(SP), AX		// copy argv
	MOVQ	AX, 8(SP)
	CALL	runtime·args(SB)
	
	/* osinit 主要用于设置 cpu 数量，见 runtime2.go: ncpu，以及设置物理页的 size */
	CALL	runtime·osinit(SB)
	
	/* 调度器初始化，详细见下文 */
	CALL	runtime·schedinit(SB)

	/* 调用 proc.go: newproc(siz int32, fn *funcval) 创建 main goroutine */
	MOVQ	$runtime·mainPC(SB), AX		// entry
	PUSHQ	AX
	PUSHQ	$0			// arg size
	CALL	runtime·newproc(SB)
	POPQ	AX
	POPQ	AX

	/* 启动 m0 */
	CALL	runtime·mstart(SB)

  /* mstart 不会返回，若返回则终止程序 */
	CALL	runtime·abort(SB)	// mstart should never return
	RET

	// Prevent dead-code elimination of debugCallV2, which is
	// intended to be called by debuggers.
	MOVQ	$runtime·debugCallV2<ABIInternal>(SB), AX
	RET

```



```go
/**************** [proc.go] ****************/

func schedinit() {
  
  ... ...
  
  /* getg() 由编译器替换为汇编指令，实际是从 TLS 中拿到当前 m 正在执行的 goroutine */
	_g_ := getg()
  
  /* 初始化 race detector 的上下文（仅当开启竞争检测时） */
	if raceenabled {
		_g_.racectx, raceprocctx0 = raceinit()
	}

  /* 调度器最多可以启动的 m 数量 */
	sched.maxmcount = 10000

	// The world starts stopped.
	worldStopped()

  /* moduledata 中存储的是与 tracing 相关的module、package、function、pc 等信息（存储在编译后的二进制文件内），如下是验证这些信息的有效性 */
	moduledataverify()
  
  /* 初始化栈 
   * 有两个全局的栈内存池：
   * 1. stackpool：存放了全局的栈 mspan 链表，可用于分配小于 32KiB 的内存空间，定义见：_StackCacheSize = 32 * 1024
   * 2. stackLarger：分配大于 32KiB 的内存
  */
	stackinit()
  
  /* 初始化堆 */
	mallocinit()
  
  /* 生成随机数，将在下面的 mcommoninit() 中用到 */
	fastrandinit() // must run before mcommoninit
  
  /* 初始化 m0
   * 并为 m0 创建一个 gsignal goroutine 用于处理系统信号，m 中的 fastrand 即前面生成的
  */
	mcommoninit(_g_.m, -1)
  
  /* 初始化 cpu，设置 cpu 扩展指令集 */
	cpuinit()       // must run before alginit
  
  /* 初始化 hash 种子 */
	alginit()       // maps must not be used before this call
  
  
	modulesinit()   // provides activeModules
	typelinksinit() // uses maps, activeModules
	itabsinit()     // uses activeModules

  /* 保存当前信号 mask */
	sigsave(&_g_.m.sigmask)
	initSigmask = _g_.m.sigmask

	if offset := unsafe.Offsetof(sched.timeToRun); offset%8 != 0 {
		println(offset)
		throw("sched.timeToRun not aligned to 8 bytes")
	}

  /* argslice 中保存 argv，envs 中保存 env，解析 debug 参数 */
	goargs()
	goenvs()
	parsedebugvars()
  
  /* 开启 GC */
	gcinit()

	lock(&sched.lock)
	sched.lastpoll = uint64(nanotime())
	procs := ncpu
	if n, ok := atoi32(gogetenv("GOMAXPROCS")); ok && n > 0 {
		procs = n
	}
  
  /* 按 GOMAXPROCS 的数量设置 p 
   * 1. 主要是设置 allp slice，并初始化其中的每一个 p
   * 2. 绑定 m0 和 p0，p0 设置为 _Prunning，其他的 p 设置为 _Pidle
  */
	if procresize(procs) != nil {
		throw("unknown runnable goroutine during bootstrap")
	}
	unlock(&sched.lock)

	// World is effectively started now, as P's can run.
	worldStarted()

	... ...
}
```



```go
/**************** [runtime2.go] ****************/

type funcval struct {
	fn uintptr
	// variable-size, fn-specific data here
}

/**************** [proc.go] ****************/

//go:nosplit
func newproc(siz int32, fn *funcval) {
  /* argp 指向 fn 函数的第一个参数 */
	argp := add(unsafe.Pointer(&fn), sys.PtrSize)
	gp := getg()
  
  /* 这里的 caller pc，指向的就是 CALL	runtime·newproc(SB) 的下一行：POPQ AX */
	pc := getcallerpc()
  
  /* systemstack 先将调用者栈切换到 g0 栈，不过目前已经在 g0 栈了，因此什么也不做 */
	systemstack(func() {
    /* 构造一个新的 g 结构，见下文 */
		newg := newproc1(fn, argp, siz, gp, pc)

    /* 目前是在 m0 执行，前文讲到 m0 绑定了 allp[0]，所以 _p_ 正是 allp[0] */
		_p_ := getg().m.p.ptr()
    
    /* 将 g 入队 */
		runqput(_p_, newg, true)

    /* 目前还没有执行 main goroutine，因此 mainStarted == false */
		if mainStarted {
			wakep()
		}
	})
}

func newproc1(fn *funcval, argp unsafe.Pointer, narg int32, callergp *g, callerpc uintptr) *g {
	... ...

	_g_ := getg()

	... ...
	
	siz := narg
	siz = (siz + 7) &^ 7

	... ...

	_p_ := _g_.m.p.ptr()
  
  /* 由于当前是在初始化第一个 goroutine，因此 gFreeList 没有空闲的 g 可用，需要创建 */
	newg := gfget(_p_)
	if newg == nil {
    /* _StackMin = 2048，因此创建一个新的 g，其栈空间为 2M */
		newg = malg(_StackMin)
		casgstatus(newg, _Gidle, _Gdead)
		allgadd(newg) // publishes with a g->status of Gdead so GC scanner doesn't look at uninitialized stack.
	}
	
  ... ...

	totalSize := 4*sys.PtrSize + uintptr(siz) + sys.MinFrameSize // extra space in case of reads slightly beyond frame
	totalSize += -totalSize & (sys.StackAlign - 1)               // align to StackAlign
	sp := newg.stack.hi - totalSize
	spArg := sp
	
  ... ...
  
	if narg > 0 {
    /* 创建 g 之前，fn 的参数是放在 caller 的栈上的，memmove 将其 copy 到 newg 的栈上 */
		memmove(unsafe.Pointer(spArg), argp, uintptr(narg))
		// This is a stack-to-stack copy. If write barriers
		// are enabled and the source stack is grey (the
		// destination is always black), then perform a
		// barrier copy. We do this *after* the memmove
		// because the destination stack may have garbage on
		// it.
		if writeBarrier.needed && !_g_.m.curg.gcscandone {
			f := findfunc(fn.fn)
			stkmap := (*stackmap)(funcdata(f, _FUNCDATA_ArgsPointerMaps))
			if stkmap.nbit > 0 {
				// We're in the prologue, so it's always stack map index 0.
				bv := stackmapdata(stkmap, 0)
				bulkBarrierBitmap(spArg, spArg, uintptr(bv.n)*sys.PtrSize, 0, bv.bytedata)
			}
		}
	}

  /* 将 newg 的 sp pc 等信息保存在 gobuf 中，待实际被调度时，就会被加载出来执行
   * 这里的 pc 存放的是 goexit + 1 的地址，这是为了让 fn 执行完毕后，跳到 goexit 来做一些退出工作，详见下文
  */
	memclrNoHeapPointers(unsafe.Pointer(&newg.sched), unsafe.Sizeof(newg.sched))
	newg.sched.sp = sp
	newg.stktopsp = sp
	newg.sched.pc = abi.FuncPCABI0(goexit) + sys.PCQuantum // +PCQuantum so that previous instruction is in same function
	newg.sched.g = guintptr(unsafe.Pointer(newg))
  
  /* 深入到 gostartcallfn 函数内我们就可以看到：
   * 该函数在 newg 的 sp 栈顶申请了一个 ptr 的位置，将 goexit 地址保存进去，然后让 sched.sp = sp-1，并将 sched.pc = fn，
   * 这实际上相当于 fake 了 fn 是由 goexit 调用的，当 fn 执行完毕后 pc 会被恢复为 goexit+1 的地址，并执行 goexit。
  */
	gostartcallfn(&newg.sched, fn)
	newg.gopc = callerpc
	newg.ancestors = saveAncestors(callergp)
	newg.startpc = fn.fn
	
  ... ...
  
  /* 修改状态为 _Grunnable，代表可以被运行了 */
	casgstatus(newg, _Gdead, _Grunnable)

	... ...

  /* 至此 main goroutine 的 g 就创建好了，返回后会进入队，并等待在 mstart 时被调度 */
	return newg
}
```



```assembly
/**************** [asm_amd64.s] ****************/

TEXT runtime·mstart(SB),NOSPLIT|TOPFRAME,$0
	CALL	runtime·mstart0(SB)
	RET // not reached
```



```go
/**************** [proc.go] ****************/

func mstart0() {
	_g_ := getg()

  /* 显然目前 _g_ == g0，所以不需要再初始化栈了 */
	osStack := _g_.stack.lo == 0
	if osStack {
		// Initialize stack bounds from system stack.
		// Cgo may have left stack size in stack.hi.
		// minit may update the stack bounds.
		//
		// Note: these bounds may not be very accurate.
		// We set hi to &size, but there are things above
		// it. The 1024 is supposed to compensate this,
		// but is somewhat arbitrary.
		size := _g_.stack.hi
		if size == 0 {
			size = 8192 * sys.StackGuardMultiplier
		}
		_g_.stack.hi = uintptr(noescape(unsafe.Pointer(&size)))
		_g_.stack.lo = _g_.stack.hi - size + 1024
	}
	// Initialize stack guard so that we can start calling regular
	// Go code.
	_g_.stackguard0 = _g_.stack.lo + _StackGuard
	// This is the g0, so we can also call go:systemstack
	// functions, which check stackguard1.
	_g_.stackguard1 = _g_.stackguard0
	
  /* 实际执行的部分，见下文 */
  mstart1()

  /* 若执行到这里，就说明主程序要结束了 */
	// Exit this thread.
	if mStackIsSystemAllocated() {
		// Windows, Solaris, illumos, Darwin, AIX and Plan 9 always system-allocate
		// the stack, but put it in _g_.stack before mstart,
		// so the logic above hasn't set osStack yet.
		osStack = true
	}
	mexit(osStack)
}

func mstart1() {
	_g_ := getg()

	if _g_ != _g_.m.g0 {
		throw("bad runtime·mstart")
	}

  /* 这里将 g0 goroutine 的调度上下文设置为跳转到前面 mstart1() 的下一句，意味着跳转后程序会结束*/
	// Set up m.g0.sched as a label returning to just
	// after the mstart1 call in mstart0 above, for use by goexit0 and mcall.
	// We're never coming back to mstart1 after we call schedule,
	// so other calls can reuse the current frame.
	// And goexit0 does a gogo that needs to return from mstart1
	// and let mstart0 exit the thread.
	_g_.sched.g = guintptr(unsafe.Pointer(_g_))
	_g_.sched.pc = getcallerpc()
	_g_.sched.sp = getcallersp()

  /* amd64 架构下是空函数 */
	asminit()
  
  /* 执行一些信号的初始化，mstartm0() 也一样 */
	minit()

	// Install signal handlers; after minit so that minit can
	// prepare the thread to be able to handle the signals.
	if _g_.m == &m0 {
		mstartm0()
	}

  /* 执行创建 m 时传入的函数，m0 没有，所以 fn == nil */
	if fn := _g_.m.mstartfn; fn != nil {
		fn()
	}

	if _g_.m != &m0 {
		acquirep(_g_.m.nextp.ptr())
		_g_.m.nextp = 0
	}
  
  /* 开始调度，经过一系列操作后，main goroutine 会被调度到 m0 上 */
	schedule()
}
```



在看 `schedule()` 之前，我们先跳到 main goroutine 的 main function 看一看：

```go
func main() {
	g := getg()
	
  ... ...
  
  /* 我们在前面 newproc 中看到了，当 mainStarted == true 时，newproc 就可以尝试创建新的 m 来执行 g 了 */
	// Allow newproc to start new Ms.
	mainStarted = true

  /* monitor 线程 */
	if GOARCH != "wasm" { // no threads on wasm yet, so no sysmon
		// For runtime_syscall_doAllThreadsSyscall, we
		// register sysmon is not ready for the world to be
		// stopped.
		atomic.Store(&sched.sysmonStarting, 1)
		systemstack(func() {
			newm(sysmon, nil, -1)
		})
	}

	... ...

  /* 执行依赖中的 init() */
	doInit(&runtime_inittask) // Must be before defer.

	... ...

  /* 启用 gc */
	gcenable()

	main_init_done = make(chan bool)
  ... ...
  /* 执行用户 main.go 中的 init() */
	doInit(&main_inittask)
	... ...
	close(main_init_done)

	needUnlock = false
	unlockOSThread()

	... ...
  
  /* 这里开始调用用户代码中的 main()，正式执行到用户代码 */
	fn := main_main // make an indirect call, as the linker doesn't know the address of the main package when laying down the runtime
	fn()

	... ...

  /* 显然当用户的 main() 执行完毕后，程序自然就可以退出了 */
	exit(0)
	for {
		var x *int32
		*x = 0
	}
}
```

