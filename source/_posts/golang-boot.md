---
title: Go 程序启动随笔
date: 2022-02-07 22:55:34
tags: 
- source
- go
categories:
- Golang
---

### 1. 启动过程

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
/**************** [proc.go] ****************/

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



```go
/**************** [proc.go] ****************/

func schedule() {
  /* 如果是从 mstart0 而来，则当前拿到的是 g0 */
	_g_ := getg()

	... ...

  /* 假如 g 所在的 m 锁定了固定运行的 goroutine，则暂停当前 m，将 m 上的 p 转移到其他 m，再运行锁定的 g*/
	if _g_.m.lockedg != 0 {
		stoplockedm()
		execute(_g_.m.lockedg.ptr(), false) // Never returns.
	}

	... ...

top:
  /* preempt == true 代表 p 需要立即进入调度，目前已经在 scheduler() 内，因此清零它 */
	pp := _g_.m.p.ptr()
	pp.preempt = false

  /* 如果当前有 GC 在等待，则先 GC，再执行调度 */
	if sched.gcwaiting != 0 {
    /* 停止当前 m，执行 GC，阻塞等待直到被唤醒，之后跳转 top，重新开始调度 */
    gcstopm()
		goto top
	}
	
  ... ...

  /* gp 就是即将被选出的 g */
	var gp *g
	var inheritTime bool

	... ...
  
  /* 为了保证公平性，当前 p 的 schedtick（每一次调度循环都 +1） 等于 61 时，强制从全局队列中拿一个 g 出来，否则如果有两个 goroutine 互相创建对方，他们就会永远占有当前 p */
	if gp == nil {
		// Check the global runnable queue once in a while to ensure fairness.
		// Otherwise two goroutines can completely occupy the local runqueue
		// by constantly respawning each other.
		if _g_.m.p.ptr().schedtick%61 == 0 && sched.runqsize > 0 {
			lock(&sched.lock)
			gp = globrunqget(_g_.m.p.ptr(), 1)
			unlock(&sched.lock)
		}
	}
  
  /* 如果 schedtick没到 61，或者全局队列也没有 g 了，就尝试从本地 runq 中获取 g */
	if gp == nil {
		gp, inheritTime = runqget(_g_.m.p.ptr())
		// We can see gp != nil here even if the M is spinning,
		// if checkTimers added a local goroutine via goready.
	}
  
  /* 如果本地 runq 里也没有 g 了，就需要通过 findrunnable() 阻塞获取 g（可能会从其他 p 的 runq 中进行工作窃取） 
   * findrunnable 会：
   * 1. 再次尝试：是否需要 gc、是否存在 finalizers g、cgo、本地 runq、全局队列等等
   * 2. 从 netpoll 中查找是否存在等待完成的 g
   * 3. 尝试工作窃取
  */
	if gp == nil {
		gp, inheritTime = findrunnable() // blocks until work is available
	}

	// This thread is going to run a goroutine and is not spinning anymore,
	// so if it was marked as spinning we need to reset it now and potentially
	// start a new spinning M.
	if _g_.m.spinning {
		resetspinning()
	}

	... ...
  
  /* 如果拿到的 g 要求必须在锁定的 m 上执行，则将之交给锁定的 m 去执行，并再次进入调度循环 */
	if gp.lockedm != 0 {
		// Hands off own p to the locked m,
		// then blocks waiting for a new p.
		startlockedm(gp)
		goto top
	}

  /* 一切就绪，准备开始调度被选中的 g 了 */
	execute(gp, inheritTime)
}

func execute(gp *g, inheritTime bool) {
	_g_ := getg()

  /* 将被调度的 g 与当前 m 绑定 */
	// Assign gp.m before entering _Grunning so running Gs have an
	// M.
	_g_.m.curg = gp
	gp.m = _g_.m
  
  /* 将状态改为 _Grunning */
	casgstatus(gp, _Grunnable, _Grunning)
  
  /* waitsince 是当前 g 被阻塞的估计时间，preempt 指示是否被抢占，重置 stackguard0 */
	gp.waitsince = 0
	gp.preempt = false
	gp.stackguard0 = gp.stack.lo + _StackGuard
	
  ... ...

  /* 传入 gobuf，跳转到汇编代码 */
	gogo(&gp.sched)
}
```



```assembly
/**************** [asm_amd64.s] ****************/

TEXT runtime·gogo(SB), NOSPLIT, $0-8
	MOVQ	buf+0(FP), BX		// gobuf
	MOVQ	gobuf_g(BX), DX
	MOVQ	0(DX), CX		// make sure g != nil
	JMP	gogo<>(SB)

TEXT gogo<>(SB), NOSPLIT, $0
	get_tls(CX)
	
	/* 恢复现场第一步：用 gobuf 中的 g，覆盖 tls 中的 g，并放入 R14 */
	MOVQ	DX, g(CX)
	MOVQ	DX, R14		// set the g register
	
	/* 恢复现场第二步：用 gobuf 中的 sp 覆盖 SP，切换到 gp 的栈 */
	MOVQ	gobuf_sp(BX), SP	// restore SP
	
	/* 恢复现场第三步：用 gobuf 中的 ret 地址覆盖 AX（amd64 下通用返回地址放在 AX） */
	MOVQ	gobuf_ret(BX), AX
	
	/* 恢复现场第四步：用 gobuf 中的 ctxt(函数调用 traceback 的上下文寄存器) 地址覆盖 DX */
	MOVQ	gobuf_ctxt(BX), DX
	
	/* 恢复现场第五步：用 gobuf 中的 bp 覆盖 BP */
	MOVQ	gobuf_bp(BX), BP
	
	/* 清空前面用过的 gobuf 值 */
	MOVQ	$0, gobuf_sp(BX)	// clear to help garbage collector
	MOVQ	$0, gobuf_ret(BX)
	MOVQ	$0, gobuf_ctxt(BX)
	MOVQ	$0, gobuf_bp(BX)
	
	/* 最后将 gobuf 中保存的 pc 写入 BX，并直接跳到 BX 处开始执行 */
	MOVQ	gobuf_pc(BX), BX
	JMP	BX
```

最后关注一下 goroutine 执行结束后的操作：

```assembly
/**************** [asm_amd64.s] ****************/

/* 本函数是在 newproc 的时候设置的 gobuf 的默认 pc，用于在 goroutine 执行结束后作为伪造调用方而跳转的 */
// The top-most function running on a goroutine
// returns to goexit+PCQuantum.
TEXT runtime·goexit(SB),NOSPLIT|NOFRAME|TOPFRAME,$0-0
	MOVD	R0, R0	// NOP
	BL	runtime·goexit1(SB)	// does not return
```

```go
/**************** [proc.go] ****************/

// Finishes execution of the current goroutine.
func goexit1() {
	if raceenabled {
		racegoend()
	}
	if trace.enabled {
		traceGoEnd()
	}
  
  /* mcall 专用做将当前执行栈切换为 g0（
   * 1. 将当前 g 的 pc、sp、bp 等保存在 gobuf
   * 2. 通过当前 g 的 m 找到 g0，切换 sp 为 g0 的sp，完成栈切换
   * 3. 调用 mcall 的传入函数 goexit0，并将切换前的 g 传入 goexit0
  */
	mcall(goexit0)
}

func goexit0(gp *g) {
  /* 这里 _g_ == g0 */
	_g_ := getg()

  /* 将原 g 状态设置为 _Gdead */
	casgstatus(gp, _Grunning, _Gdead)

  ... ...
  
  /* 做一些原 g 的清理工作 */
	gp.m = nil
	locked := gp.lockedm != 0
	gp.lockedm = 0
	_g_.m.lockedg = 0
	gp.preemptStop = false
	gp.paniconfault = false
	gp._defer = nil // should be true already but just in case.
	gp._panic = nil // non-nil for Goexit during panic. points at stack-allocated data.
	gp.writebuf = nil
	gp.waitreason = 0
	gp.param = nil
	gp.labels = nil
	gp.timer = nil

	... ...

  /* 将 m 与 curg 的关联断开*/
	dropg()

	... ...
  
  /* 原 g 执行完了，将其剩余的部分放入 gfree list，以便复用 */
	gfput(_g_.m.p.ptr(), gp)
	
  ... ...
  
  /* 重新进入调度循环 */
	schedule()
}
```



### 2. M 系统线程操作

m 是作为程序的实际执行载体，首先看看创建 m：

```go
/**************** [proc.go] ****************/

func newm(fn func(), _p_ *p, id int64) {
  /* 创建 m 结构 */
	mp := allocm(_p_, fn, id)
  
  /* 执行 p 的 m 可以 park；设置 nextp 为 _p_；设置 sigmask */
	mp.doesPark = (_p_ != nil)
	mp.nextp.set(_p_)
	mp.sigmask = initSigmask
	
  ... ...
	
  /* 创建系统线程 */
  newm1(mp)
}

func allocm(_p_ *p, fn func(), id int64) *m {
	_g_ := getg()
	acquirem() // disable GC because it can be called from sysmon
	if _g_.m.p == 0 {
		acquirep(_p_) // temporarily borrow p for mallocs in this function
	}

	... ...

  /* 初始化 m 结构*/
	mp := new(m)
	mp.mstartfn = fn
	mcommoninit(mp, id)

  /* 每个 m 都有自己的 g0，初始化 g0 */
	// In case of cgo or Solaris or illumos or Darwin, pthread_create will make us a stack.
	// Windows and Plan 9 will layout sched stack on OS stack.
	if iscgo || mStackIsSystemAllocated() {
		mp.g0 = malg(-1)
	} else {
		mp.g0 = malg(8192 * sys.StackGuardMultiplier)
	}
	mp.g0.m = mp

	if _p_ == _g_.m.p.ptr() {
		releasep()
	}
	releasem(_g_.m)

	return mp
}

func newm1(mp *m) {
	... ...
	execLock.rlock() // Prevent process clone.
  /* 根据不同操作系统，按照实际系统创建系统线程 */
	newosproc(mp)
	execLock.runlock()
}
```

```go
/**************** [os_linux.go] ****************/

func newosproc(mp *m) {
	stk := unsafe.Pointer(mp.g0.stack.hi)
	/*
	 * note: strace gets confused if we use CLONE_PTRACE here.
	 */
	if false {
		print("newosproc stk=", stk, " m=", mp, " g=", mp.g0, " clone=", funcPC(clone), " id=", mp.id, " ostk=", &mp, "\n")
	}

	// Disable signals during clone, so that the new thread starts
	// with signals disabled. It will enable them in minit.
	var oset sigset
	sigprocmask(_SIG_SETMASK, &sigset_all, &oset)
  
  /* 通过系统调用 clone 创建 linux 线程 */
	ret := clone(cloneFlags, stk, unsafe.Pointer(mp), unsafe.Pointer(mp.g0), unsafe.Pointer(funcPC(mstart)))
	
  sigprocmask(_SIG_SETMASK, &oset, nil)

	if ret < 0 {
		print("runtime: failed to create new OS thread (have ", mcount(), " already; errno=", -ret, ")\n")
		if ret == -_EAGAIN {
			println("runtime: may need to increase max user processes (ulimit -u)")
		}
		throw("newosproc")
	}
}

```

```assembly
/**************** [sys_linux_amd64.s] ****************/

// int32 clone(int32 flags, void *stk, M *mp, G *gp, void (*fn)(void));
TEXT runtime·clone(SB),NOSPLIT,$0
  /* 在 os_linux.go 中可以查到传入的 flags：*/
    cloneFlags = _CLONE_VM | /* share memory */
		_CLONE_FS | /* share cwd, etc */
		_CLONE_FILES | /* share fd table */
		_CLONE_SIGHAND | /* share sig handler table */
		_CLONE_SYSVSEM | /* share SysV semaphore undo lists (see issue #20763) */
		_CLONE_THREAD /* revisit - okay for now */
  /* 更多详情：https://man7.org/linux/man-pages/man2/clone.2.html */
	MOVL	flags+0(FP), DI
	
	/* 这里传入的是 g0 的 stack.hi */
	MOVQ	stk+8(FP), SI
	MOVQ	$0, DX
	MOVQ	$0, R10
	MOVQ    $0, R8
	
	/* mp gp fn 等结构原本是在父线程栈内创建的，需要 copy 到新线程栈内 */
	// Copy mp, gp, fn off parent stack for use by child.
	// Careful: Linux system call clobbers CX and R11.
	MOVQ	mp+16(FP), R13
	MOVQ	gp+24(FP), R9
	MOVQ	fn+32(FP), R12
	CMPQ	R13, $0    // m
	JEQ	nog1
	CMPQ	R9, $0    // g
	JEQ	nog1
	
	/* 前面 m、g 都不为 0，因此保存 m_tls 到 R8 */
	LEAQ	m_tls(R13), R8
#ifdef GOOS_android
	// Android stores the TLS offset in runtime·tls_g.
	SUBQ	runtime·tls_g(SB), R8
#else
	ADDQ	$8, R8	// ELF wants to use -8(FS)
#endif
	ORQ 	$0x00080000, DI //add flag CLONE_SETTLS(0x00080000) to call clone
nog1:
  /* call clone 系统调用 */
	MOVL	$SYS_clone, AX
	SYSCALL

  /* 由于 clone 创建了新的线程空间，对于子线程，返回值 AX = 0 代表创建成功，对于父线程，返回值 AX 放入的是子线程 pid */
	// In parent, return.
	CMPQ	AX, $0
	JEQ	3(PC)
	
	/* 这里是父线程，copy pid 到栈上，直接返回 */
	MOVL	AX, ret+40(FP)
	RET

  /* 这里是子线程，先恢复 g0 栈 */
	// In child, on new stack.
	MOVQ	SI, SP

	// If g or m are nil, skip Go-related setup.
	CMPQ	R13, $0    // m
	JEQ	nog2
	CMPQ	R9, $0    // g
	JEQ	nog2

  /* 获取当前线程 id，放入 m_procid */
	// Initialize m->procid to Linux tid
	MOVL	$SYS_gettid, AX
	SYSCALL
	MOVQ	AX, m_procid(R13)

  /* 恢复 m，g */
	// In child, set up new stack
	get_tls(CX)
	MOVQ	R13, g_m(R9)
	MOVQ	R9, g(CX)
	MOVQ	R9, R14 // set g register
	CALL	runtime·stackcheck(SB)

nog2:
  /* 执行传入的 fn，根据 newosproc，这里执行的是 mstart，这里又回到了启动时候的路径，mstart 的终点，就是 schedule() */
	// Call fn. This is the PC of an ABI0 function.
	CALL	R12

  /* 正常情况下 schdule() 是永远不返回的，如果返回了，就关闭当前线程 */
	// It shouldn't return. If it does, exit that thread.
	MOVL	$111, DI
	MOVL	$SYS_exit, AX
	SYSCALL
	JMP	-3(PC)	// keep exiting
```


```go
/**************** [proc.go] ****************/

/* startm 用于将传入的 _p_ 放到 m 上执行 */
func startm(_p_ *p, spinning bool) {
  /* 先给当前 m 加锁，若传入的 _p_ 为 nil 则从 pidle 中获取空闲的 p，没有空闲的 p 则退出 */
	mp := acquirem()
	lock(&sched.lock)
	if _p_ == nil {
		_p_ = pidleget()
		if _p_ == nil {
			unlock(&sched.lock)
			if spinning {
				// The caller incremented nmspinning, but there are no idle Ps,
				// so it's okay to just undo the increment and give up.
				if int32(atomic.Xadd(&sched.nmspinning, -1)) < 0 {
					throw("startm: negative nmspinning")
				}
			}
			releasem(mp)
			return
		}
	}
  
  /* 尝试从 midle 中获取一个空闲的 m，若获取不到，就创建一个新的 m */
	nmp := mget()
	if nmp == nil {
    /* 先为 m 分配一个 id，防止在释放 sched.lock 后运行 checkdead() 时被判定为死锁 */
		id := mReserveID()
		unlock(&sched.lock)

		var fn func()
    /* 如果期望启动一个 spining 状态的 m，那么新创建的 m 就是正在自旋的 */
		if spinning {
			// The caller incremented nmspinning, so set m.spinning in the new M.
			fn = mspinning
		}
		newm(fn, _p_, id)
		// Ownership transfer of _p_ committed by start in newm.
		// Preemption is now safe.
		releasem(mp)
		return
	}
	unlock(&sched.lock)
  
  /* 从 midle 中拿到的 m 一定不处于自旋状态（只有 stopm 后，m 才会进入 midle，而停止的 m 不处于自旋态） */
	if nmp.spinning {
		throw("startm: m is spinning")
	}
  
  /* midle 中的 m 不应拥有 p */
	if nmp.nextp != 0 {
		throw("startm: m has p")
	}
  
  /* 如果传入了非空闲的 p，且还期望启动一个自旋的 m，是自相矛盾的 */
	if spinning && !runqempty(_p_) {
		throw("startm: p has runnable gs")
	}
  
  /* 根据需要自旋的情况设置自旋 */
	// The caller incremented nmspinning, so set m.spinning in the new M.
	nmp.spinning = spinning
	nmp.nextp.set(_p_)
  
  /* 唤醒该 m 所绑定的系统线程，正式开始工作
   *（如果是从 findrunnable() 调用的 stopm()，那么就会继续执行 findrunnable() 寻找新的 g） 
  */
	notewakeup(&nmp.park)
	// Ownership transfer of _p_ committed by wakeup. Preemption is now
	// safe.
	releasem(mp)
}

/* 在 m 确实找不到可运行的 g 时，将被放入 midle 中，同时其关联的系统线程也将休眠 */
func stopm() {
	_g_ := getg()

	if _g_.m.locks != 0 {
		throw("stopm holding locks")
	}
	if _g_.m.p != 0 {
		throw("stopm holding p")
	}
	if _g_.m.spinning {
		throw("stopm spinning")
	}

	lock(&sched.lock)
  /* midle 入队 */
	mput(_g_.m)
	unlock(&sched.lock)
  
  /* 系统线程休眠 */
	mPark()
	acquirep(_g_.m.nextp.ptr())
	_g_.m.nextp = 0
}
```

```go
/**************** [lock_futex.go] ****************/

/* 在 linux 系统下，notesleep 和 notewakeup 是基于 futex 实现，而在 macos 下则是 mutex cond */
func notesleep(n *note) {
	gp := getg()
	if gp != gp.m.g0 {
		throw("notesleep not on g0")
	}
	ns := int64(-1)
	if *cgo_yield != nil {
		// Sleep for an arbitrary-but-moderate interval to poll libc interceptors.
		ns = 10e6
	}
	for atomic.Load(key32(&n.key)) == 0 {
		gp.m.blocked = true
		futexsleep(key32(&n.key), 0, ns)
		if *cgo_yield != nil {
			asmcgocall(*cgo_yield, nil)
		}
		gp.m.blocked = false
	}
}

func notewakeup(n *note) {
	old := atomic.Xchg(key32(&n.key), 1)
	if old != 0 {
		print("notewakeup - double wakeup (", old, ")\n")
		throw("notewakeup - double wakeup")
	}
	futexwakeup(key32(&n.key), 1)
}
```



### 3. 栈

```go
/**************** [stack.go] ****************/

/* 为调用者分配一段栈空间，返回栈顶与栈底地址（即 stack 结构） */
func stackalloc(n uint32) stack {
  /* 需要在系统栈上运行，因此 thisg 一定是 g0 */
	thisg := getg()

  ... ...
  
	// Small stacks are allocated with a fixed-size free-list allocator.
	// If we need a stack of a bigger size, we fall back on allocating
	// a dedicated span.
	var v unsafe.Pointer
  /* 1. _FixedStack 是根据 _StackSystem 所调整的 _StackMin 的二次幂数值，
   *    在 linux 下 _StackSystem = 0，因此 linux 下 _FixedStack == _StackMin == 2048
   * 2. _NumStackOrders 代表栈阶数，golang 中将栈分为 n 阶，第 k + 1 阶的栈容量是 k 阶的 2 倍
   *    linux 下 _NumStackOrders = 4，其每阶栈空间大小分别是 2k、4k、8k、16k
   * 3. _StackCacheSize = 32k，小于 32k 的栈空间分配可以尝试在 P 缓存中分配
  */
	if n < _FixedStack<<_NumStackOrders && n < _StackCacheSize {
		order := uint8(0)
		n2 := n
    /* 根据需要分配的空间大小 n，找到需要在哪一阶（order）中分配 */
		for n2 > _FixedStack {
			order++
			n2 >>= 1
		}
		var x gclinkptr
		if stackNoCache != 0 || thisg.m.p == 0 || thisg.m.preemptoff != "" {
      /* 部分情况下需要直接从全局的 stackpool 中分配栈空间，stackpoolalloc 见后文 */
			// thisg.m.p == 0 can happen in the guts of exitsyscall
			// or procresize. Just get a stack from the global pool.
			// Also don't touch stackcache during gc
			// as it's flushed concurrently.
			lock(&stackpool[order].item.mu)
			x = stackpoolalloc(order)
			unlock(&stackpool[order].item.mu)
		} else {
      /* 否则就可以在 P 缓存内分配栈空间
       * 每个 p 结构中都持有一个 mcache，其中保存了 _NumStackOrders 个 stackfreelist 用来存放空闲栈空间
      */
			c := thisg.m.p.ptr().mcache
			x = c.stackcache[order].list
			if x.ptr() == nil {
        /* 若当前阶的 stackfreelist 还未分配，对其进行分配
         *（实际上还是从全局 stackpool 中分配，一次性分配 _StackCacheSize 的一半）
         * 分配的栈空间会以一个个的 segment 的形式链式存储，每个 segment 的容量等于当前阶所规定的栈容量
        */
				stackcacherefill(c, order)
				x = c.stackcache[order].list
			}
      /* 从空闲列表 head 处分配 n 个字节，并将当前空闲列表 head 指向下一个 segment，若指到了 tail，下一次就又会对其进行分配
       * 由于 _StackMin = 2k，又要求 n 必须是二次幂，因此就能确保待分配的 n 字节栈空间恰好与对应阶的栈容量一致
      */
			c.stackcache[order].list = x.ptr().next
			c.stackcache[order].size -= uintptr(n)
		}
		v = unsafe.Pointer(x)
	} else {
    /* 若需要的栈空间过大，就尝试在 stackLarge 中分配
     * stackLarge 实际上存放的是一组 span 链表，每一条链表存放的占空间大小是以一页大小（8k）的二次幂划分的，
     * 即 stackLarge.free[0] =》 8k，stackLarge.free[1] =》 16k，stackLarge.free[1] =》 32k，... ...
     * 同样的，由于 _StackMin = 2k，又要求 n 必须是二次幂，因此传入的 n 转换为 npages 后也会遵循 1，2，4，8... 的要求
    */
		var s *mspan
    /* 计算需要的 page 数量 */
		npage := uintptr(n) >> _PageShift
		log2npage := stacklog2(npage)

		// Try to get a stack from the large stack cache.
		lock(&stackLarge.lock)
		if !stackLarge.free[log2npage].isEmpty() {
      /* 如果存在合适的空闲空间，直接使用 */
			s = stackLarge.free[log2npage].first
			stackLarge.free[log2npage].remove(s)
		}
		unlock(&stackLarge.lock)

		lockWithRankMayAcquire(&mheap_.lock, lockRankMheap)

    /* 如果 stackLarge 中没有剩余空间了，那么直接从堆中分配 npage 空间 */
		if s == nil {
			// Allocate a new stack from the heap.
			s = mheap_.allocManual(npage, spanAllocStack)
			if s == nil {
				throw("out of memory")
			}
			osStackAlloc(s)
			s.elemsize = uintptr(n)
		}
		v = unsafe.Pointer(s.base())
	}

	... ...
	return stack{uintptr(v), uintptr(v) + uintptr(n)}
}

/* 全局 stackpool 的分配 */
func stackpoolalloc(order uint8) gclinkptr {
  /* 每一阶都存放了一个 span 链表作为空闲链表 */
	list := &stackpool[order].item.span
	s := list.first
	lockWithRankMayAcquire(&mheap_.lock, lockRankMheap)
  /* 如果当前阶的空闲列表为空，尝试从堆分配，但 allocManual 所分配的内存是不被 gc 管理的，因此一定要手动释放 */
	if s == nil {
    /* 一次性分配 _StackCacheSize 容量的页，返回一个 mspan */
		// no free stacks. Allocate another span worth.
		s = mheap_.allocManual(_StackCacheSize>>_PageShift, spanAllocStack)
		
    ... ...
		
    /* 按 order 所指定的大小，将内存切分成 segments，用 manualFreeList 管理 */
    s.elemsize = _FixedStack << order
		for i := uintptr(0); i < _StackCacheSize; i += s.elemsize {
			x := gclinkptr(s.base() + i)
			x.ptr().next = s.manualFreeList
			s.manualFreeList = x
		}
		list.insert(s)
	}
	x := s.manualFreeList
	if x.ptr() == nil {
		throw("span has no free stacks")
	}
	s.manualFreeList = x.ptr().next
	s.allocCount++
  /* 当 manualFreeList 为 nil 时说明当前 mspan 持有的空闲 segment 都已经被分配掉了，因此把它移除 stackpool */
	if s.manualFreeList.ptr() == nil {
		// all stacks in s are allocated.
		list.remove(s)
	}
  /* 最终返回的是 mspan manualFreeList 的首地址 */
	return x
}

/* 将 stack 释放，通常在释放 g 的时候调用 */
func stackfree(stk stack) {
	gp := getg()
	v := unsafe.Pointer(stk.lo)
	n := stk.hi - stk.lo

  ... ...
  
  /* 一样是根据栈空间大小来决定释放的位置 */
	if n < _FixedStack<<_NumStackOrders && n < _StackCacheSize {
		order := uint8(0)
		n2 := n
		for n2 > _FixedStack {
			order++
			n2 >>= 1
		}
		x := gclinkptr(v)
    /* 选择从 stackpool 或者 stackcache 中释放 */
		if stackNoCache != 0 || gp.m.p == 0 || gp.m.preemptoff != "" {
			lock(&stackpool[order].item.mu)
			stackpoolfree(x, order)
			unlock(&stackpool[order].item.mu)
		} else {
			c := gp.m.p.ptr().mcache
			if c.stackcache[order].size >= _StackCacheSize {
        /* 当 stackcache 已满时，尝试释放一部分（保证容量小于 _StackCacheSize/2） */
				stackcacherelease(c, order)
			}
      /* 把 stack segment 插入 head */
			x.ptr().next = c.stackcache[order].list
			c.stackcache[order].list = x
			c.stackcache[order].size += n
		}
	} else {
    /* 若栈空间过大 */
		s := spanOfUnchecked(uintptr(v))
		if s.state.get() != mSpanManual {
			println(hex(s.base()), v)
			throw("bad span state")
		}
    /* 假如当前 gc 未运行（在后台 sweep）直接将空间释放 */
		if gcphase == _GCoff {
			// Free the stack immediately if we're
			// sweeping.
			osStackFree(s)
			mheap_.freeManual(s, spanAllocStack)
		} else {
      /* gc 正在运行时，不能直接释放，还回 stackLarge */
			// If the GC is running, we can't return a
			// stack span to the heap because it could be
			// reused as a heap span, and this state
			// change would race with GC. Add it to the
			// large stack cache instead.
			log2npage := stacklog2(s.npages)
			lock(&stackLarge.lock)
			stackLarge.free[log2npage].insert(s)
			unlock(&stackLarge.lock)
		}
	}
}

/* 释放全局 stackpool */
func stackpoolfree(x gclinkptr, order uint8) {
	s := spanOfUnchecked(uintptr(x))
  /* 用作 stack 的 mspan，都会持有 mSpanManual 状态 */
	if s.state.get() != mSpanManual {
		throw("freeing stack not in a stack span")
	}
  /* 假如 manualFreeList 为 nil，把它重新加回 stackpool，因为在释放了空间后，这个 span 会重新拥有空闲空间 */
	if s.manualFreeList.ptr() == nil {
		// s will now have a free stack
		stackpool[order].item.span.insert(s)
	}
	x.ptr().next = s.manualFreeList
	s.manualFreeList = x
	s.allocCount--
  /* 当 gc 在后台未运行，且当前 span 的所有 segment 都被释放了，就把他还回 heap */
	if gcphase == _GCoff && s.allocCount == 0 {
		// Span is completely free. Return it to the heap
		// immediately if we're sweeping.
		//
		// If GC is active, we delay the free until the end of
		// GC to avoid the following type of situation:
		//
		// 1) GC starts, scans a SudoG but does not yet mark the SudoG.elem pointer
		// 2) The stack that pointer points to is copied
		// 3) The old stack is freed
		// 4) The containing span is marked free
		// 5) GC attempts to mark the SudoG.elem pointer. The
		//    marking fails because the pointer looks like a
		//    pointer into a free span.
		//
		// By not freeing, we prevent step #4 until GC is done.
		stackpool[order].item.span.remove(s)
		s.manualFreeList = 0
		osStackFree(s)
		mheap_.freeManual(s, spanAllocStack)
	}
}
```

```assembly
/**************** [asm_amd64.s] ****************/

/* 检查是否需要扩容当前栈
 * 对于每一个非 NOSPLIT 的函数，编译器都会在最前面插入尝试调用 morestack 的逻辑：
 * 若当前 SP 已经小于 stackgourd0，则跳转到 morestack
*/
TEXT runtime·morestack(SB),NOSPLIT,$0-0
	// Cannot grow scheduler stack (m->g0).
	get_tls(CX)
	MOVQ	g(CX), BX
	MOVQ	g_m(BX), BX
	MOVQ	m_g0(BX), SI
	/*  g0 的栈不能被扩容，因此如果检测到 morestack 在 g0 上被调用，直接终止程序 */
	CMPQ	g(CX), SI
	JNE	3(PC)
	CALL	runtime·badmorestackg0(SB)
	CALL	runtime·abort(SB)

  /*  gsignal 的栈也不能被扩容 */
	// Cannot grow signal stack (m->gsignal).
	MOVQ	m_gsignal(BX), SI
	CMPQ	g(CX), SI
	JNE	3(PC)
	CALL	runtime·badmorestackgsignal(SB)
	CALL	runtime·abort(SB)

  /* 将当前函数调用者的 pc sp g 等信息保存在 m 的 morebuf 中 */
	// Called from f.
	// Set m->morebuf to f's caller.
	NOP	SP	// tell vet SP changed - stop checking offsets
	MOVQ	8(SP), AX	// f's caller's PC
	MOVQ	AX, (m_morebuf+gobuf_pc)(BX)
	LEAQ	16(SP), AX	// f's caller's SP
	MOVQ	AX, (m_morebuf+gobuf_sp)(BX)
	get_tls(CX)
	MOVQ	g(CX), SI
	MOVQ	SI, (m_morebuf+gobuf_g)(BX)

  /* 将当前函数的 pc sp g 等信息保存在 g 的 shecd 中 */
	// Set g->sched to context in f.
	MOVQ	0(SP), AX // f's PC
	MOVQ	AX, (g_sched+gobuf_pc)(SI)
	LEAQ	8(SP), AX // f's SP
	MOVQ	AX, (g_sched+gobuf_sp)(SI)
	MOVQ	BP, (g_sched+gobuf_bp)(SI)
	MOVQ	DX, (g_sched+gobuf_ctxt)(SI)

  /* 切换到 g0 运行 newstack */
	// Call newstack on m->g0's stack.
	MOVQ	m_g0(BX), BX
	MOVQ	BX, g(CX)
	MOVQ	(g_sched+gobuf_sp)(BX), SP
	CALL	runtime·newstack(SB)
	/* newstack 会直接调用 gogo 跳转到原 goroutine 执行，因此不会返回 */
	CALL	runtime·abort(SB)	// crash if newstack returns
	RET
```

```go
/**************** [stack.go] ****************/

func newstack() {
	thisg := getg()
  
  ... ...

	gp := thisg.m.curg

  ... ...
  
	morebuf := thisg.m.morebuf
	thisg.m.morebuf.pc = 0
	thisg.m.morebuf.lr = 0
	thisg.m.morebuf.sp = 0
	thisg.m.morebuf.g = 0

  /* 这里略过了抢占相关的逻辑，当下只关心栈扩容 */
	... ...

  /* 新栈空间是旧栈的两倍 */
	// Allocate a bigger segment and move the stack.
	oldsize := gp.stack.hi - gp.stack.lo
	newsize := oldsize * 2

  /* 通过 PCDATA 计算出函数所需的栈帧空间，如果新栈所扩张的空间仍然不够函数所需，则对他再次乘 2 */
	// Make sure we grow at least as much as needed to fit the new frame.
	// (This is just an optimization - the caller of morestack will
	// recheck the bounds on return.)
	if f := findfunc(gp.sched.pc); f.valid() {
		max := uintptr(funcMaxSPDelta(f))
		needed := max + _StackGuard
		used := gp.stack.hi - gp.sched.sp
		for newsize-used < needed {
			newsize *= 2
		}
	}

	... ...
  
  /* 假如扩容后的栈空间超过了最大容量，抛出栈溢出错误，maxstacksize 会在 main goroutine 中被设置为 1 GiB */
	if newsize > maxstacksize || newsize > maxstackceiling {
		if maxstacksize < maxstackceiling {
			print("runtime: goroutine stack exceeds ", maxstacksize, "-byte limit\n")
		} else {
			print("runtime: goroutine stack exceeds ", maxstackceiling, "-byte limit\n")
		}
		print("runtime: sp=", hex(sp), " stack=[", hex(gp.stack.lo), ", ", hex(gp.stack.hi), "]\n")
		throw("stack overflow")
	}

	// The goroutine must be executing in order to call newstack,
	// so it must be Grunning (or Gscanrunning).
	casgstatus(gp, _Grunning, _Gcopystack)

	// The concurrent GC will not scan the stack while we are doing the copy since
	// the gp is in a Gcopystack status.
	copystack(gp, newsize)
	
  if stackDebug >= 1 {
		print("stack grow done\n")
	}
	casgstatus(gp, _Gcopystack, _Grunning)
	gogo(&gp.sched)
}

/* 扩容实际上是分配新空间，并将旧栈内容复制进去 */
func copystack(gp *g, newsize uintptr) {
  ... ...
	old := gp.stack
  ... ...
	used := old.hi - gp.sched.sp

  /* 用 stackalloc 分配新空间 */
	// allocate new stack
	new := stackalloc(uint32(newsize))
	... ...
  
  /* 计算新旧栈空间之间的距离，便于后续 copy 的时候定位 */
	// Compute adjustment.
	var adjinfo adjustinfo
	adjinfo.old = old
	adjinfo.delta = new.hi - old.hi

	// Adjust sudogs, synchronizing with channel ops if necessary.
	ncopy := used
  /* 对于没有非阻塞 channel 指向当前 g 的 stack，直接调整其每一个 sudog 指向的 stack 位置
   * （sudog 即 pseudo g，sudog 代表 g 在等待某个对象，g.waiting 是当前 g 的等待队列）
  */
	if !gp.activeStackChans {
		if newsize < old.hi-old.lo && atomic.Load8(&gp.parkingOnChan) != 0 {
			// It's not safe for someone to shrink this stack while we're actively
			// parking on a channel, but it is safe to grow since we do that
			// ourselves and explicitly don't want to synchronize with channels
			// since we could self-deadlock.
			throw("racy sudog adjustment due to parking on channel")
		}
		adjustsudogs(gp, &adjinfo)
	} else {
    /* 否则，需要将 channel lock 之后再移动 */
		// sudogs may be pointing in to the stack and gp has
		// released channel locks, so other goroutines could
		// be writing to gp's stack. Find the highest such
		// pointer so we can handle everything there and below
		// carefully. (This shouldn't be far from the bottom
		// of the stack, so there's little cost in handling
		// everything below it carefully.)
		adjinfo.sghi = findsghi(gp, old)

		// Synchronize with channel ops and copy the part of
		// the stack they may interact with.
		ncopy -= syncadjustsudogs(gp, used, &adjinfo)
	}

  /* memmove 位于 memmove_amd64.s ，实现中有非常多优化 */
	// Copy the stack (or the rest of it) to the new location
	memmove(unsafe.Pointer(new.hi-ncopy), unsafe.Pointer(old.hi-ncopy), ncopy)

  /* 在 sched、defer、panic 中与 stack 相关的全都要调整 */
	// Adjust remaining structures that have pointers into stacks.
	// We have to do most of these before we traceback the new
	// stack because gentraceback uses them.
	adjustctxt(gp, &adjinfo)
	adjustdefers(gp, &adjinfo)
	adjustpanics(gp, &adjinfo)
	if adjinfo.sghi != 0 {
		adjinfo.sghi += adjinfo.delta
	}

  /* 将 g.stack 切换 */
	// Swap out old stack for new one
	gp.stack = new
	gp.stackguard0 = new.lo + _StackGuard // NOTE: might clobber a preempt request
	gp.sched.sp = new.hi - used
	gp.stktopsp += adjinfo.delta

	// Adjust pointers in the new stack.
	gentraceback(^uintptr(0), ^uintptr(0), 0, gp, 0, nil, 0x7fffffff, adjustframe, noescape(unsafe.Pointer(&adjinfo)), 0)

  /* 最后把旧空间释放 */
	// free old stack
	if stackPoisonCopy != 0 {
		fillstack(old, 0xfc)
	}
	stackfree(old)
}

/* 缩小栈 */
func shrinkstack(gp *g) {
	... ...

  /* 准备将栈空间缩小为原来的一半，但必须大于 _FixedStack，否则不缩小 */
	oldsize := gp.stack.hi - gp.stack.lo
	newsize := oldsize / 2
	// Don't shrink the allocation below the minimum-sized stack
	// allocation.
	if newsize < _FixedStack {
		return
	}
  
  /* 如果已使用的栈空间大于总空间的四分之一，也不缩小 */
	// Compute how much of the stack is currently in use and only
	// shrink the stack if gp is using less than a quarter of its
	// current stack. The currently used stack includes everything
	// down to the SP plus the stack guard space that ensures
	// there's room for nosplit functions.
	avail := gp.stack.hi - gp.stack.lo
	if used := gp.stack.hi - gp.sched.sp + _StackLimit; used >= avail/4 {
		return
	}

	if stackDebug > 0 {
		print("shrinking stack ", oldsize, "->", newsize, "\n")
	}

  /* 缩小的逻辑还是通过将栈内容复制到一个更小的空间内完成的 */
	copystack(gp, newsize)
}
```



### 4. 堆

```go
/**************** [malloc.go] ****************/

func newobject(typ *_type) unsafe.Pointer {
	return mallocgc(typ.size, typ, true)
}

func mallocgc(size uintptr, typ *_type, needzero bool) unsafe.Pointer {
	/* 这里略过了 GODEBUG = sbrk，gcAssist 协助标记等逻辑，当下只关心堆内存分配 */
  ... ...

	// Set mp.mallocing to keep from being preempted by GC.
	mp := acquirem()
	... ...
	mp.mallocing = 1

	shouldhelpgc := false
	dataSize := size
	c := getMCache()
  ... ...
  
	var span *mspan
	var x unsafe.Pointer
	noscan := typ == nil || typ.ptrdata == 0
	// In some cases block zeroing can profitably (for latency reduction purposes)
	// be delayed till preemption is possible; isZeroed tracks that state.
	isZeroed := true
  
  /* 首先判断需要分配的空间是否大于 maxSmallSize（=32KiB），大对象会进入专门的分配逻辑 */
	if size <= maxSmallSize {
    /* 当传入的 type 为 nil 或者不是指针类型（typ.ptrdata == 0），且所需容量小于 maxTinySize（16 byte）时，
     * 使用 tiny allocator 
    */
		if noscan && size < maxTinySize {
      /* Tiny Allocator
       * 将许多小对象组合起来，共同分配一块空间（maxTinySize）。
       * maxTinySize 目前是 16 bytes，最多可能浪费一倍的空间。
       * tiny allocator 主要用于分配小字符串和单个逃逸的变量，tiny allocator 的空间在 mcache 中分配。
      */
			off := c.tinyoffset
			/* 先对当前 tiny block 的 offset 按照所需的 size 进行对齐 */
			if size&7 == 0 {
				off = alignUp(off, 8)
			} else if sys.PtrSize == 4 && size == 12 {
				off = alignUp(off, 8)
			} else if size&3 == 0 {
				off = alignUp(off, 4)
			} else if size&1 == 0 {
				off = alignUp(off, 2)
			}
      
      /* 对齐后的 offset + size 仍旧小于 maxTinySize，说明可以直接在当前块分配 */
			if off+size <= maxTinySize && c.tiny != 0 {
				// The object fits into existing tiny block.
				x = unsafe.Pointer(c.tiny + off)
				c.tinyoffset = off + size
        /* 记录了总分配数 */
				c.tinyAllocs++
				mp.mallocing = 0
				releasem(mp)
        
        /* 返回内存块起始地址，结束 */
				return x
			}
			
      /* 若当前块不能满足所需空间，就需要新创建一个 maxTinySize 大的新块（旧块中的剩余空间浪费掉了） */
			span = c.alloc[tinySpanClass]
      /* 先看看在 mcache 里面有没有空闲内存 */
			v := nextFreeFast(span)
			if v == 0 {
        /* 如果没有了，就需要分配一个新的 span（具体见内存管理部分） */
				v, span, shouldhelpgc = c.nextFree(tinySpanClass)
			}
			x = unsafe.Pointer(v)
      /* maxTinySize = 16 bytes，清空 */
			(*[2]uint64)(x)[0] = 0
			(*[2]uint64)(x)[1] = 0
			// See if we need to replace the existing tiny block with the new one
			// based on amount of remaining free space.
			if !raceenabled && (size < c.tinyoffset || c.tiny == 0) {
				// Note: disabled when race detector is on, see comment near end of this function.
				c.tiny = uintptr(x)
				c.tinyoffset = size
			}
			size = maxTinySize
		} else {
      /* 大于 16 bytes 或者是指针类型 */
			var sizeclass uint8
			if size <= smallSizeMax-8 {
				sizeclass = size_to_class8[divRoundUp(size, smallSizeDiv)]
			} else {
				sizeclass = size_to_class128[divRoundUp(size-smallSizeMax, largeSizeDiv)]
			}
      
      /* 根据计算过的 sizeclass 来选择分配的空间大小（而不是直接分配实际需要的容量） */
			size = uintptr(class_to_size[sizeclass])
			spc := makeSpanClass(sizeclass, noscan)
			span = c.alloc[spc]
      
      /* 一样，有空闲就在 mcache 找，没有就再分配 */
			v := nextFreeFast(span)
			if v == 0 {
				v, span, shouldhelpgc = c.nextFree(spc)
			}
			x = unsafe.Pointer(v)
			if needzero && span.needzero != 0 {
        /* 若需要，对分配的空间进行清空 */
				memclrNoHeapPointers(unsafe.Pointer(v), size)
			}
		}
	} else {
    /* 大对象，直接在堆上分配 */
		shouldhelpgc = true
		// For large allocations, keep track of zeroed state so that
		// bulk zeroing can be happen later in a preemptible context.
		span, isZeroed = c.allocLarge(size, needzero && !noscan, noscan)
		span.freeindex = 1
		span.allocCount = 1
		x = unsafe.Pointer(span.base())
		size = span.elemsize
	}

  /* 省略了 GC、调试等等逻辑 */
	... ...

	return x
}

```



### 5. 内存管理

#### 5.1 内存分配

```
         ┌────────────────┐
         │     mcache     │
         ├────────────────┤
         │    mcentral    │
         ├────────────────┤
         │     mheap      │
      ┌──┴──┬─────┬─────┬─┴───┐
      │ span│ span│ span│ span│
    ┌─┼─┬─┬─┼─────┴────┬┴┬─┬─┬┴┬─┐
    │ │ │ │ │ heapArena│ │ │ │ │ │
    ├─┴─┴─┴─┴──────────┴─┴─┴─┴─┴─┤
    │         os memory          │
    └────────────────────────────┘
```

```go
/**************** [mcache.go] ****************/

/* 前面提到过的 mcache，由每个 p 持有，避免了内存操作的锁竞争 */
type mcache struct {
	/* 主要用与 mem profile */
	nextSample uintptr // trigger heap sample after allocating this many bytes
  /* mcache alloc 申请的内存部分需要 gc scan，这里是容量计数 */
	scanAlloc  uintptr // bytes of scannable heap allocated

  /* 前文分配堆内存中讲到的 Tiny Allocator
   * tiny = 起始地址
   * tinyoffset = 当前 span 的空闲位置
   * tinyAllocs = 已分配对象计数
  */
	tiny       uintptr
	tinyoffset uintptr
	tinyAllocs uintptr

  /* 小于 _MaxSmallSize = 32768 的对象都在这里分配
   * numSpanClasses = 68 << 1 分别包含 68 个 scan 的 span 和 68 个 noscan 的 span
  */
	alloc [numSpanClasses]*mspan // spans to allocate from, indexed by spanClass

  /* 栈空间在这里分配，按照 _NumStackOrders，分配栈的大小分别是 2K 4K 8K 16K  */
	stackcache [_NumStackOrders]stackfreelist

	/* 当前 mcache 最后一次 GC flush 时的 sweep generation*/
	flushGen uint32
}

/* mcache 本身也需要内存空间来存放，这里在 g0 栈上给 mcache 分配空间，并创造 mcache */
func allocmcache() *mcache {
	var c *mcache
	systemstack(func() {
		lock(&mheap_.lock)
    /* 主要逻辑是通过 mheap_.cachealloc 来分配，见后文 */
		c = (*mcache)(mheap_.cachealloc.alloc())
		c.flushGen = mheap_.sweepgen
		unlock(&mheap_.lock)
	})
	for i := range c.alloc {
		c.alloc[i] = &emptymspan
	}
	c.nextSample = nextSample()
	return c
}

/* 释放逻辑类似 */
func freemcache(c *mcache) {
	systemstack(func() {
		c.releaseAll()
		stackcache_clear(c)

		lock(&mheap_.lock)
		mheap_.cachealloc.free(unsafe.Pointer(c))
		unlock(&mheap_.lock)
	})
}

/* 分配新的 span */
func (c *mcache) nextFree(spc spanClass) (v gclinkptr, s *mspan, shouldhelpgc bool) {
	/* 根据 spanClass 选取合适的 span */
  s = c.alloc[spc]
	shouldhelpgc = false
  
  /* 获取当前 span 的空闲对象位置 */
	freeIndex := s.nextFreeIndex()
  /* 若没有空闲对象了，需要分配新的 */
	if freeIndex == s.nelems {
    ... ...
    /* 从 mcentral 中申请新的 span */
		c.refill(spc)
		shouldhelpgc = true
		s = c.alloc[spc]
    
    /* 获取新 span 的 freeIndex */
		freeIndex = s.nextFreeIndex()
	}

  ... ...

	v = gclinkptr(freeIndex*s.elemsize + s.base())
	s.allocCount++
  
	... ...
  
	return
}

func (c *mcache) refill(spc spanClass) {
	// Return the current cached span to the central lists.
	s := c.alloc[spc]

	if uintptr(s.allocCount) != s.nelems {
		throw("refill of span with free space remaining")
	}
  
  /* 刚初始化的 span 都是 emptymspan，如果当前获取到的 span 不是 emptymspan，就把他先还回 mcentral */
	if s != &emptymspan {
		// Mark this span as no longer cached.
		if s.sweepgen != mheap_.sweepgen+3 {
			throw("bad sweepgen in refill")
		}
		mheap_.central[spc].mcentral.uncacheSpan(s)
	}

  /* 从 mcentral 获取新的 span，从这里能看出，每一个 span class 对应了一个 mcantral */
	s = mheap_.central[spc].mcentral.cacheSpan()
	
  ... ...

  /* 对获取到的 span 进行处理 */
	// Indicate that this span is cached and prevent asynchronous
	// sweeping in the next sweep phase.
	s.sweepgen = mheap_.sweepgen + 3

	// Assume all objects from this span will be allocated in the
	// mcache. If it gets uncached, we'll adjust this.
	stats := memstats.heapStats.acquire()
	atomic.Xadduintptr(&stats.smallAllocCount[spc.sizeclass()], uintptr(s.nelems)-uintptr(s.allocCount))

	// Flush tinyAllocs.
	if spc == tinySpanClass {
		atomic.Xadduintptr(&stats.tinyAllocCount, c.tinyAllocs)
		c.tinyAllocs = 0
	}
	memstats.heapStats.release()

	// Update gcController.heapLive with the same assumption.
	usedBytes := uintptr(s.allocCount) * s.elemsize
	atomic.Xadd64(&gcController.heapLive, int64(s.npages*pageSize)-int64(usedBytes))

	// While we're here, flush scanAlloc, since we have to call
	// revise anyway.
	atomic.Xadd64(&gcController.heapScan, int64(c.scanAlloc))
	c.scanAlloc = 0

	if trace.enabled {
		// gcController.heapLive changed.
		traceHeapAlloc()
	}
	if gcBlackenEnabled != 0 {
		// gcController.heapLive and heapScan changed.
		gcController.revise()
	}

  /* 最后加入对应位置 */
	c.alloc[spc] = s
}

/* 大对象直接进入此方法分配 */
func (c *mcache) allocLarge(size uintptr, needzero bool, noscan bool) (*mspan, bool) {
	/* 距离到达地址指针最大值不足一页 */
  if size+_PageSize < size {
		throw("out of memory")
	}
  
  /* 计算需要的整数页 */
	npages := size >> _PageShift
	if size&_PageMask != 0 {
		npages++
	}

	// Deduct credit for this span allocation and sweep if
	// necessary. mHeap_Alloc will also sweep npages, so this only
	// pays the debt down to npage pages.
	deductSweepCredit(npages*_PageSize, npages)

  /* 根据是否 noscan 获取正确的 span class，注意此处的 size class 是 0，代表分配的是大于 32k 的大内存 */
	spc := makeSpanClass(0, noscan)
  
  /* 直接从堆分配 npages 页内存，由于 size class = 0，因此不受 class_to_size 和 class_to_allocnpages 的限制 */
	s, isZeroed := mheap_.alloc(npages, spc, needzero)
	if s == nil {
		throw("out of memory")
	}
	stats := memstats.heapStats.acquire()
	atomic.Xadduintptr(&stats.largeAlloc, npages*pageSize)
	atomic.Xadduintptr(&stats.largeAllocCount, 1)
	memstats.heapStats.release()

	// Update gcController.heapLive and revise pacing if needed.
	atomic.Xadd64(&gcController.heapLive, int64(npages*pageSize))
	if trace.enabled {
		// Trace that a heap alloc occurred because gcController.heapLive changed.
		traceHeapAlloc()
	}
	if gcBlackenEnabled != 0 {
		gcController.revise()
	}

	// Put the large span in the mcentral swept list so that it's
	// visible to the background sweeper.
	mheap_.central[spc].mcentral.fullSwept(mheap_.sweepgen).push(s)
	s.limit = s.base() + size
	heapBitsForAddr(s.base()).initSpan(s)
	return s, isZeroed
}
```



```go
/**************** [mcentral.go] ****************/

type mcentral struct {
  /* 当前 mcentral 所属的 span class */
	spanclass spanClass

  /*
   * partial 和 full 各包含两个 mspan set：一个是已经清理的 span，另一个是未清理的 span。
   * 每次 GC 后，已清理的和未清理的 spanSet 会互换。未清理的 spanSet 会被内存分配器或是 GC 后台清理器抽取。
   * 
   * 每次 GC，sweepgen 会被 +2，所以已清理的 spanSet = partial[sweepgen/2%2]，
   * 未清理的 spanSet = partial[1 - sweepgen/2%2]
  */
	partial [2]spanSet // list of spans with a free object
	full    [2]spanSet // list of spans with no free objects
}

/* 向 mcache 提供 span 的具体方法 */
func (c *mcentral) cacheSpan() *mspan {
	... ...

  /* 不论是在部分空闲还是无空闲列表中尝试超过 spanBudget 次，还没有找到合适的 span，
   * 就直接分配一个新的 span，以此减小对小对象清理的开销 
  */
	spanBudget := 100

	var s *mspan
	sl := newSweepLocker()
	sg := sl.sweepGen

	/* 先尝试从已清理的部分空闲集合中获取 span */
	if s = c.partialSwept(sg).pop(); s != nil {
		goto havespan
	}

	/* 如果没有，就从未清理部分空闲集合中获取 span */
	for ; spanBudget >= 0; spanBudget-- {
		s = c.partialUnswept(sg).pop()
		if s == nil {
			break
		}
		if s, ok := sl.tryAcquire(s); ok {
			/* 锁定了一个未清理的 span，将其清理后使用 */
			s.sweep(true)
			sl.dispose()
			goto havespan
		}
		/* 假如没能锁定当前的未清理 span，说明它已经被后台异步清理器锁定了，但还没有来得及将其处理完并移除出未清理列表，找下一个 */
	}
	
  /* 如果还是没有，尝试从未清理无空闲列表中获取 span */
	for ; spanBudget >= 0; spanBudget-- {
		s = c.fullUnswept(sg).pop()
		if s == nil {
			break
		}
		if s, ok := sl.tryAcquire(s); ok {
			/* 还是先清理 */
			s.sweep(true)
			/* 之后看看是不是空出了空闲位置 */
			freeIndex := s.nextFreeIndex()
			if freeIndex != s.nelems {
				s.freeindex = freeIndex
				sl.dispose()
				goto havespan
			}
			/* 若没找到任何空闲位置，把它放入已扫描无空闲列表中，重试下一个 */
			c.fullSwept(sg).push(s.mspan)
		}
		// See comment for partial unswept spans.
	}
	sl.dispose()
	if trace.enabled {
		traceGCSweepDone()
		traceDone = true
	}

	/* 实在找不到可使用的现存 span 了，向 heap 申请新的，假如还申请不到就只好 OOM */
	s = c.grow()
	if s == nil {
		return nil
	}

	/* 程序执行到此处，证明一定找到一个有空闲的 span 了 */
havespan:
	if trace.enabled && !traceDone {
		traceGCSweepDone()
	}
	n := int(s.nelems) - int(s.allocCount)
	if n == 0 || s.freeindex == s.nelems || uintptr(s.allocCount) == s.nelems {
		throw("span has no free objects")
	}
	freeByteBase := s.freeindex &^ (64 - 1)
	whichByte := freeByteBase / 8
	// Init alloc bits cache.
	s.refillAllocCache(whichByte)

	// Adjust the allocCache so that s.freeindex corresponds to the low bit in
	// s.allocCache.
	s.allocCache >>= s.freeindex % 64

	return s
}

/* 从 heap 中申请新 span */
func (c *mcentral) grow() *mspan {
	npages := uintptr(class_to_allocnpages[c.spanclass.sizeclass()])
	size := uintptr(class_to_size[c.spanclass.sizeclass()])

	s, _ := mheap_.alloc(npages, c.spanclass, true)
	if s == nil {
		return nil
	}

	// Use division by multiplication and shifts to quickly compute:
	// n := (npages << _PageShift) / size
	n := s.divideByElemSize(npages << _PageShift)
	s.limit = s.base() + size*n
	heapBitsForAddr(s.base()).initSpan(s)
	return s
}

/* 从 mcache 回收 span 
 * span 的 sweepgen 与全局 sweepgen 的关系：
 * span sweepgen == 全局 sweepgen - 2：需要清理
 * span sweepgen == 全局 sweepgen - 1：正被清理
 * span sweepgen == 全局 sweepgen：清理完成，且可用
 * span sweepgen == 全局 sweepgen + 1：在清理前就被 cache 了，需要清理
 * span sweepgen == 全局 sweepgen + 3：清理完成且仍旧是 cache 状态
 * 全局 sweepgen 每次 GC 自动 +2
*/
func (c *mcentral) uncacheSpan(s *mspan) {
	if s.allocCount == 0 {
		throw("uncaching span but s.allocCount == 0")
	}

	sg := mheap_.sweepgen
  /* 若等于 sg+1 代表需要清理 */
	stale := s.sweepgen == sg+1

	// Fix up sweepgen.
	if stale {
		/* 去除 cache 状态，且正在清理 */
		atomic.Store(&s.sweepgen, sg-1)
	} else {
		/* 去除 cache 状态 */
		atomic.Store(&s.sweepgen, sg)
	}

	// Put the span in the appropriate place.
	if stale {
    /* stale 的 span 不在全局清理列表中，可以直接锁定 */
		ss := sweepLocked{s}
    
    /* 清理，由于传入参数 preserve == false，所以清理完后就归还给 heap 或 spanSet */
		ss.sweep(false)
	} else {
		if int(s.nelems)-int(s.allocCount) > 0 {
			/* 还有空余，归还到已清理部分空闲列表 */
			c.partialSwept(sg).push(s)
		} else {
			/* 没有空余了，归还到已清理无空闲列表 */
			c.fullSwept(sg).push(s)
		}
	}
}
```

```go
/**************** [mheap.go] ****************/

type mheap struct {
	lock  mutex
  /* 页分配器 */
	pages pageAlloc // page allocation data structure

  /* 清理相关 */
	sweepgen     uint32 // sweep generation, see comment in mspan; written during STW
	sweepDrained uint32 // all spans are swept or are being swept
	sweepers     uint32 // number of active sweepone calls

  /* 所有被创建出来的 span */
	allspans []*mspan // all spans out there

	_ uint32 // align uint64 fields on 32-bit for atomics

	/* 扫描清理相关参数 */
	pagesInUse         uint64  // pages of spans in stats mSpanInUse; updated atomically
	pagesSwept         uint64  // pages swept this cycle; updated atomically
	pagesSweptBasis    uint64  // pagesSwept to use as the origin of the sweep ratio; updated atomically
	sweepHeapLiveBasis uint64  // value of gcController.heapLive to use as the origin of sweep ratio; written with lock, read without
	sweepPagesPerByte  float64 // proportional sweep ratio; written with lock, read without

  /* runtime 保留的将要还给 os 的内存数量 */
	scavengeGoal uint64
  
	// This is accessed atomically.
  /* 页回收状态
   * reclaimIndex：指向 allArenas 中下一个待回收的页
   * reclaimCredit：清理出的比所需空间更多的空间，计数并放入 reclaimCredit
  */
	reclaimIndex uint64
	reclaimCredit uintptr

  /* arena map，每个 arena 帧管理着一块虚拟地址空间
   * heap 中未分配的空间，arena 指向 nil
   * 为了节省 arena 帧的数量，可能会存在多级 arena map，但在多数的 64 位平台上，只有一级
  */
	arenas [1 << arenaL1Bits]*[1 << arenaL2Bits]*heapArena

  /* 用于存放 heapArena map，防止与 heap 本身产生交错 */
	heapArenaAlloc linearAlloc
  
	arenaHints *arenaHint

	/* 用于存放 arena 本身 */
	arena linearAlloc

	/* 所有已经分配的 arena index，用做基于地址空间迭代 */
	allArenas []arenaIdx

	/* 在清理周期开始前对 allArenas 的快照 */
	sweepArenas []arenaIdx

	/* 在标记周期开始前对 allArenas 的快照 */
	markArenas []arenaIdx

	/* 当前 heap 生长到的 arena */
	curArena struct {
		base, end uintptr
	}

	_ uint32 // ensure 64-bit alignment of central

	/* mcentral 列表，68*2，分别包含 scan 与 noscan 两类 */
	central [numSpanClasses]struct {
		mcentral mcentral
		pad      [cpu.CacheLinePadSize - unsafe.Sizeof(mcentral{})%cpu.CacheLinePadSize]byte
	}

  /* 各种固定大小对象内存分配器 */
	spanalloc             fixalloc // allocator for span*
	cachealloc            fixalloc // allocator for mcache*
	specialfinalizeralloc fixalloc // allocator for specialfinalizer*
	specialprofilealloc   fixalloc // allocator for specialprofile*
	specialReachableAlloc fixalloc // allocator for specialReachable
	speciallock           mutex    // lock for special record allocators.
	arenaHintAlloc        fixalloc // allocator for arenaHints

	unused *specialfinalizer // never set, just here to force the specialfinalizer type into DWARF
}

/* 初始化 heap */
func (h *mheap) init() {
  ... ...
  
  /* 初始化各种 Fixed Size 对象分配器，init 中的 first 是钩子函数，会在每一次分配时调用*/
	h.spanalloc.init(unsafe.Sizeof(mspan{}), recordspan, unsafe.Pointer(h), &memstats.mspan_sys)
	h.cachealloc.init(unsafe.Sizeof(mcache{}), nil, nil, &memstats.mcache_sys)
	h.specialfinalizeralloc.init(unsafe.Sizeof(specialfinalizer{}), nil, nil, &memstats.other_sys)
	h.specialprofilealloc.init(unsafe.Sizeof(specialprofile{}), nil, nil, &memstats.other_sys)
	h.specialReachableAlloc.init(unsafe.Sizeof(specialReachable{}), nil, nil, &memstats.other_sys)
	h.arenaHintAlloc.init(unsafe.Sizeof(arenaHint{}), nil, nil, &memstats.other_sys)

  /* 分配的 span 不需要 0 初始化 */
	h.spanalloc.zero = false

  /* 初始化 mcentral */
	for i := range h.central {
		h.central[i].mcentral.init(spanClass(i))
	}

  /* 初始化页分配器 */
	h.pages.init(&h.lock, &memstats.gcMiscSys)
}

/* 堆内存分配入口 */
func (h *mheap) alloc(npages uintptr, spanclass spanClass, needzero bool) (*mspan, bool) {
	var s *mspan
  /* 必须在系统栈上操作 heap，否则可能会触发栈扩容，而栈扩容本身可能会导致调用本方法 */
	systemstack(func() {
		/* 为了防止 heap 过多的被分配，在分配空间之前，先尝试回收至少 npages 空间
		 *（如果 isSweepDone == true 证明所有 span 都扫描过了，也就不需要再尝试回收） 
		*/
		if !isSweepDone() {
      /* 尝试回收 npages 页内存 */
			h.reclaim(npages)
		}
    
    /* 按需要分配 span */
		s = h.allocSpan(npages, spanAllocHeap, spanclass)
	})

	if s == nil {
		return nil, false
	}
	isZeroed := s.needzero == 0
	if needzero && !isZeroed {
    /* 内存清零 */
		memclrNoHeapPointers(unsafe.Pointer(s.base()), s.npages<<_PageShift)
		isZeroed = true
	}
	s.needzero = 0
	return s, isZeroed
}

/* 回收内存 */
func (h *mheap) reclaim(npage uintptr) {
	... ...

	arenas := h.sweepArenas
	locked := false
	for npage > 0 {
		/* 前文提到 reclaimCredit 代表每次回收内存时多回收的页数，因此此处先从 reclaimCredit 中扣减 */
		if credit := atomic.Loaduintptr(&h.reclaimCredit); credit > 0 {
			take := credit
			if take > npage {
				// Take only what we need.
				take = npage
			}
			if atomic.Casuintptr(&h.reclaimCredit, credit, credit-take) {
				npage -= take
			}
			continue
		}

    /* 从 reclaimIndex 获取需要回收的 chunk（512 个页大小的块） 的起始 id*/
		// Claim a chunk of work.
		idx := uintptr(atomic.Xadd64(&h.reclaimIndex, pagesPerReclaimerChunk) - pagesPerReclaimerChunk)
		if idx/pagesPerArena >= uintptr(len(arenas)) {
			// Page reclaiming is done.
			atomic.Store64(&h.reclaimIndex, 1<<63)
			break
		}

		if !locked {
			// Lock the heap for reclaimChunk.
			lock(&h.lock)
			locked = true
		}

		/* 扫描并回收，范围是 pagesPerReclaimerChunk 个页 */
		nfound := h.reclaimChunk(arenas, idx, pagesPerReclaimerChunk)
		if nfound <= npage {
			npage -= nfound
		} else {
			/* 若回收了多于 npages 的页，将其计数累加到 reclaimCredit */
			atomic.Xadduintptr(&h.reclaimCredit, nfound-npage)
			npage = 0
		}
	}
	
  ... ...
}

/* 扫描并回收 */
func (h *mheap) reclaimChunk(arenas []arenaIdx, pageIdx, n uintptr) uintptr {
	... ...
  
	for n > 0 {
		ai := arenas[pageIdx/pagesPerArena]
    
    /* 找到管理起始页的 arena */
		ha := h.arenas[ai.l1()][ai.l2()]

		/* 获取起始页在当前 arena 内的相对位置（位图） */
		arenaPage := uint(pageIdx % pagesPerArena)
    
    /* 计算从起始页位开始的所有 pageInUse 和 pageMarks 位，过长则截断 */
		inUse := ha.pageInUse[arenaPage/8:]
		marked := ha.pageMarks[arenaPage/8:]
		if uintptr(len(inUse)) > n/8 {
			inUse = inUse[:n/8]
			marked = marked[:n/8]
		}

		/* 查找这个 chunk 内正在使用且没有被标记对象（inUseUnmarked）的 span */
		for i := range inUse {
			/* 当前 inUse[i] 所指示的 8 个 span，都不符合要求 */
      inUseUnmarked := atomic.Load8(&inUse[i]) &^ marked[i]
			if inUseUnmarked == 0 {
				continue
			}

			for j := uint(0); j < 8; j++ {
				if inUseUnmarked&(1<<j) != 0 {
					s := ha.spans[arenaPage+uint(i)*8+j]
					if s, ok := sl.tryAcquire(s); ok {
						npages := s.npages
						unlock(&h.lock)
            /* 找到当前指示的 span，尝试清除 */
						if s.sweep(false) {
							nFreed += npages
						}
						lock(&h.lock)
						// Reload inUse. It's possible nearby
						// spans were freed when we dropped the
						// lock and we don't want to get stale
						// pointers from the spans array.
						inUseUnmarked = atomic.Load8(&inUse[i]) &^ marked[i]
					}
				}
			}
		}

		// Advance.
		pageIdx += uintptr(len(inUse) * 8)
		n -= uintptr(len(inUse) * 8)
	}
	sl.dispose()
	... ...
	return nFreed
}

/* 分配 span */
func (h *mheap) allocSpan(npages uintptr, typ spanAllocType, spanclass spanClass) (s *mspan) {
	... ...

	/* 对于小于四分之一 pageCachePages 的分配请求，优先从每一个 p 的 pageCache 中分配
   * pageCachePages = 8 * unsafe.Sizeof(pageCache{}.cache) = 64，cache 是 uint64 类型的 bitmap
   */
	pp := gp.m.p.ptr()
	if !needPhysPageAlign && pp != nil && npages < pageCachePages/4 {
		c := &pp.pcache

		/* 若没有任何空闲空间了，则重新给 pageCache 分配空间，注意如果页分配器发现没有空闲空间了，会返回一个空的 pageCache 结构 */
		if c.empty() {
			lock(&h.lock)
			*c = h.pages.allocToCache()
			unlock(&h.lock)
		}

		/* 从 pageCache 中尝试寻找空闲空间 */
		base, scav = c.alloc(npages)
    /* base 不为零，说明成功找到了空间 */
		if base != 0 {
      /* 尝试从 p 的 mspan cache 中获取 span 结构 */
			s = h.tryAllocMSpan()
			if s != nil {
				goto HaveSpan
			}
		}
	}

  /* p 中的 pageCache 可以并发获取，但逻辑走到这里，就必须锁整个 heap */
  lock(&h.lock)
  
	... ...
  
  /* 从上面可以看到，如果是页分配器没有空间了，base 为零，因此需要新分配 */
	if base == 0 {
		/* 尝试直接从页分配器中寻找 npages 空间 */
		base, scav = h.pages.alloc(npages)
		if base == 0 {
      /* 还是没空间，这时候必须要从 os 真正的申请新内存了 */
			if !h.grow(npages) {
				unlock(&h.lock)
				return nil
			}
      
      /* 现在再次查找，一定能找到，否则就是 os 内存不足 */
			base, scav = h.pages.alloc(npages)
			if base == 0 {
				throw("grew heap, but no adequate free space found")
			}
		}
	}
  
	if s == nil {
		/* 没有 span，就创建一个新的来（创建的同时也放入了 p 的 spancache） */
		s = h.allocMSpanLocked()
	}

	... ...

  /* 与 heap 相关的操作结束了，释放锁 */
	unlock(&h.lock)

HaveSpan:
	/* 下面是将 span 进行初始化，包括 base 地址等等 */
	s.init(base, npages)
	if h.allocNeedsZero(base, npages) {
		s.needzero = 1
	}
	nbytes := npages * pageSize
	if typ.manual() {
    /* 栈空间分配与 span class 无关 */
		s.manualFreeList = 0
		s.nelems = 0
		s.limit = s.base() + s.npages*pageSize
		s.state.set(mSpanManual)
	} else {
    /* 按照 span class 初始化其他相关属性 */
		s.spanclass = spanclass
		if sizeclass := spanclass.sizeclass(); sizeclass == 0 {
			s.elemsize = nbytes
			s.nelems = 1
			s.divMul = 0
		} else {
			s.elemsize = uintptr(class_to_size[sizeclass])
			s.nelems = nbytes / s.elemsize
			s.divMul = class_to_divmagic[sizeclass]
		}

		// Initialize mark and allocation structures.
		s.freeindex = 0
		s.allocCache = ^uint64(0) // all 1s indicating all free.
		s.gcmarkBits = newMarkBits(s.nelems)
		s.allocBits = newAllocBits(s.nelems)

		// It's safe to access h.sweepgen without the heap lock because it's
		// only ever updated with the world stopped and we run on the
		// systemstack which blocks a STW transition.
		atomic.Store(&s.sweepgen, h.sweepgen)

		// Now that the span is filled in, set its state. This
		// is a publication barrier for the other fields in
		// the span. While valid pointers into this span
		// should never be visible until the span is returned,
		// if the garbage collector finds an invalid pointer,
		// access to the span may race with initialization of
		// the span. We resolve this race by atomically
		// setting the state after the span is fully
		// initialized, and atomically checking the state in
		// any situation where a pointer is suspect.
		s.state.set(mSpanInUse)
	}

	// Commit and account for any scavenged memory that the span now owns.
	if scav != 0 {
		// sysUsed all the pages that are actually available
		// in the span since some of them might be scavenged.
		sysUsed(unsafe.Pointer(base), nbytes)
		atomic.Xadd64(&memstats.heap_released, -int64(scav))
	}
	// Update stats.
	if typ == spanAllocHeap {
		atomic.Xadd64(&memstats.heap_inuse, int64(nbytes))
	}
	if typ.manual() {
		// Manually managed memory doesn't count toward heap_sys.
		memstats.heap_sys.add(-int64(nbytes))
	}
	// Update consistent stats.
	stats := memstats.heapStats.acquire()
	atomic.Xaddint64(&stats.committed, int64(scav))
	atomic.Xaddint64(&stats.released, -int64(scav))
	switch typ {
	case spanAllocHeap:
		atomic.Xaddint64(&stats.inHeap, int64(nbytes))
	case spanAllocStack:
		atomic.Xaddint64(&stats.inStacks, int64(nbytes))
	case spanAllocPtrScalarBits:
		atomic.Xaddint64(&stats.inPtrScalarBits, int64(nbytes))
	case spanAllocWorkBuf:
		atomic.Xaddint64(&stats.inWorkBufs, int64(nbytes))
	}
	memstats.heapStats.release()

	/* 将 span 加入对应的 arena */
	h.setSpans(s.base(), npages, s)

  /* 将 span 加入对应的 pageInUse 中 */
	if !typ.manual() {
		// Mark in-use span in arena page bitmap.
		//
		// This publishes the span to the page sweeper, so
		// it's imperative that the span be completely initialized
		// prior to this line.
		arena, pageIdx, pageMask := pageIndexOf(s.base())
		atomic.Or8(&arena.pageInUse[pageIdx], pageMask)

		// Update related page sweeper stats.
		atomic.Xadd64(&h.pagesInUse, int64(npages))
	}

	// Make sure the newly allocated span will be observed
	// by the GC before pointers into the span are published.
	publicationBarrier()

	return s
}

/* 释放 span */
func (h *mheap) freeSpanLocked(s *mspan, typ spanAllocType) {
	... ...
  
  /* 在页分配器处标记空闲 */
	h.pages.free(s.base(), s.npages)

	... ...
  
  /* span 结构也释放掉 */
	h.freeMSpanLocked(s)
}

/* 尝试扩张新内存 */
func (h *mheap) grow(npage uintptr) bool {
	/* grow 需要在加锁状态 */
  assertLockHeld(&h.lock)

	/* 按 chunk 所管理的页数整数对齐 */
	ask := alignUp(npage, pallocChunkPages) * pageSize

	totalGrowth := uintptr(0)
	// This may overflow because ask could be very large
	// and is otherwise unrelated to h.curArena.base.
	end := h.curArena.base + ask
	nBase := alignUp(end, physPageSize)
	if nBase > h.curArena.end || /* overflow */ end < h.curArena.base {
		/* 当前的 arena 放不下需要扩张的空间，因此必须重新申请新的 arena */
		av, asize := h.sysAlloc(ask)
		if av == nil {
			print("runtime: out of memory: cannot allocate ", ask, "-byte block (", memstats.heap_sys, " in use)\n")
			return false
		}

		if uintptr(av) == h.curArena.end {
			/* 如果新分配的空间起始地址等于 curArena 的结束地址，说明分配了连续内存，直接扩展 curArena */
			h.curArena.end = uintptr(av) + asize
		} else {
			/* 若不连续，需要把 curArena 切到新申请的空间，而原 arena 空间需要释放给 */
			if size := h.curArena.end - h.curArena.base; size != 0 {
				/* Reserved -> Prepared 以备后用 */
				sysMap(unsafe.Pointer(h.curArena.base), size, &memstats.heap_sys)
				... ...
        
        /* 把这段空间发布给页分配器 */
				h.pages.grow(h.curArena.base, size)
				totalGrowth += size
			}
			/* 切到新的 arean 上 */
			h.curArena.base = uintptr(av)
			h.curArena.end = uintptr(av) + asize
		}

		... ...
	}

	// Grow into the current arena.
	v := h.curArena.base
	h.curArena.base = nBase

  // 新空间 Reserved -> Prepared.
	sysMap(unsafe.Pointer(v), nBase-v, &memstats.heap_sys)
  
	... ...
  
  /* 实际占用的空间是从 v 开始，大小为 nBase-v 的空间区域，更新页分配器使这部分空间可被分配 */
	h.pages.grow(v, nBase-v)
	totalGrowth += nBase - v

	// We just caused a heap growth, so scavenge down what will soon be used.
	// By scavenging inline we deal with the failure to allocate out of
	// memory fragments by scavenging the memory fragments that are least
	// likely to be re-used.
	if retained := heapRetained(); retained+uint64(totalGrowth) > h.scavengeGoal {
		todo := totalGrowth
		if overage := uintptr(retained + uint64(totalGrowth) - h.scavengeGoal); todo > overage {
			todo = overage
		}
		h.pages.scavenge(todo, false)
	}
	return true
}

func (h *mheap) sysAlloc(n uintptr) (v unsafe.Pointer, size uintptr) {
	... ...
  
  /* 只在 32 位平台生效：先从预留空间中尝试分配，预留空间 arena 会在 mallocinit 时被初始化 */
	v = h.arena.alloc(n, heapArenaBytes, &memstats.heap_sys)
	if v != nil {
		size = n
		goto mapped
	}

	/* 通过 arenaHint 申请 os 内存 */
	for h.arenaHints != nil {
		hint := h.arenaHints
		p := hint.addr
		if hint.down {
			p -= n
		}
		if p+n < p {
			// We can't use this, so don't ask.
			v = nil
		} else if arenaIndex(p+n-1) >= 1<<arenaBits {
			// Outside addressable heap. Can't use.
			v = nil
		} else {
			v = sysReserve(unsafe.Pointer(p), n)
		}
    
    /* 如果 os 返回的内存地址与 hint 中计算出的一致，申请成功 */
		if p == uintptr(v) {
			// Success. Update the hint.
			if !hint.down {
				p += n
			}
			hint.addr = p
			size = n
			break
		}
    
		/* 申请不成功则尝试下一个 hint，并释放当前 hint */
		if v != nil {
			sysFree(v, n, nil)
		}
		h.arenaHints = hint.next
		h.arenaHintAlloc.free(unsafe.Pointer(hint))
	}

  /* 所有的 hint 都不管用了，直接向 os 申请新空间，并创建新的 hint */
	if size == 0 {
		... ...
    
		v, size = sysReserveAligned(nil, n, heapArenaBytes)
		if v == nil {
			return nil, 0
		}

		hint := (*arenaHint)(h.arenaHintAlloc.alloc())
		hint.addr, hint.down = uintptr(v), true
		hint.next, mheap_.arenaHints = mheap_.arenaHints, hint
		hint = (*arenaHint)(h.arenaHintAlloc.alloc())
		hint.addr = uintptr(v) + size
		hint.next, mheap_.arenaHints = mheap_.arenaHints, hint
	}

	... ...
  
mapped:
	/* 对分配的内存创建 arena */
	for ri := arenaIndex(uintptr(v)); ri <= arenaIndex(uintptr(v)+size-1); ri++ {
		l2 := h.arenas[ri.l1()]
		if l2 == nil {
			/* 除了 64bit Windows 平台外，L1 都等于 1，L2 不存在则说明整个 arena map 未创建，因此创建之  */
			l2 = (*[1 << arenaL2Bits]*heapArena)(persistentalloc(unsafe.Sizeof(*l2), sys.PtrSize, nil))
			if l2 == nil {
				throw("out of memory allocating heap arena map")
			}
			atomic.StorepNoWB(unsafe.Pointer(&h.arenas[ri.l1()]), unsafe.Pointer(l2))
		}

    /* 上层 caller 在未找到 arean 时才会调用本方法，因此 arena 一定不存在 */
		if l2[ri.l2()] != nil {
			throw("arena already initialized")
		}
    
    /* 创建 arena */
		var r *heapArena
		r = (*heapArena)(h.heapArenaAlloc.alloc(unsafe.Sizeof(*r), sys.PtrSize, &memstats.gcMiscSys))
		if r == nil {
			r = (*heapArena)(persistentalloc(unsafe.Sizeof(*r), sys.PtrSize, &memstats.gcMiscSys))
			if r == nil {
				throw("out of memory allocating heap arena metadata")
			}
		}

		/* 将新的 arena 加到 allArena 列表后面 */
		if len(h.allArenas) == cap(h.allArenas) {
			size := 2 * uintptr(cap(h.allArenas)) * sys.PtrSize
			if size == 0 {
				size = physPageSize
			}
			newArray := (*notInHeap)(persistentalloc(size, sys.PtrSize, &memstats.gcMiscSys))
			if newArray == nil {
				throw("out of memory allocating allArenas")
			}
			oldSlice := h.allArenas
			*(*notInHeapSlice)(unsafe.Pointer(&h.allArenas)) = notInHeapSlice{newArray, len(h.allArenas), int(size / sys.PtrSize)}
			copy(h.allArenas, oldSlice)
			// Do not free the old backing array because
			// there may be concurrent readers. Since we
			// double the array each time, this can lead
			// to at most 2x waste.
		}
		h.allArenas = h.allArenas[:len(h.allArenas)+1]
		h.allArenas[len(h.allArenas)-1] = ri

		// Store atomically just in case an object from the
		// new heap arena becomes visible before the heap lock
		// is released (which shouldn't happen, but there's
		// little downside to this).
		atomic.StorepNoWB(unsafe.Pointer(&l2[ri.l2()]), unsafe.Pointer(r))
	}

	... ...

	return
}

/* 内存管理单元 */
type mspan struct {
	next *mspan     // next span in list, or nil if none
	prev *mspan     // previous span in list, or nil if none
	list *mSpanList // For debugging. TODO: Remove.

  /* 起始地址与管理页数 */
	startAddr uintptr // address of first byte of span aka s.base()
	npages    uintptr // number of pages in span

	manualFreeList gclinkptr // list of free objects in mSpanManual spans

	/* 根据 span class，每一种 span 的容量以及可分配对象数是固定的，因此：
   * freeindex：下一个空闲对象槽位
   * nelems：总对象数，如果 freeindex == nelems 则证明 span 已满
  */
	freeindex uintptr
	nelems uintptr // number of object in the span.

	/* allocBits 的补码，方便快速通过 ctz (count trailing zero) 方法快速查找空闲位置 */
	allocCache uint64

  /* 内存占用和 gc 的位图标记 */
	allocBits  *gcBits
	gcmarkBits *gcBits

	// sweep generation:
	// if sweepgen == h->sweepgen - 2, the span needs sweeping
	// if sweepgen == h->sweepgen - 1, the span is currently being swept
	// if sweepgen == h->sweepgen, the span is swept and ready to use
	// if sweepgen == h->sweepgen + 1, the span was cached before sweep began and is still cached, and needs sweeping
	// if sweepgen == h->sweepgen + 3, the span was swept and then cached and is still cached
	// h->sweepgen is incremented by 2 after every GC

	sweepgen    uint32
	divMul      uint32        // for divide by elemsize
	allocCount  uint16        // number of allocated objects
	spanclass   spanClass     // size class and noscan (uint8)
	state       mSpanStateBox // mSpanInUse etc; accessed atomically (get/set methods)
	needzero    uint8         // needs to be zeroed before allocation
	elemsize    uintptr       // computed from sizeclass or from npages
	limit       uintptr       // end of data in span
	speciallock mutex         // guards specials list
	specials    *special      // linked list of special records sorted by offset.
}
```

```
操作系统内存管理抽象层：
1) None - Unreserved and unmapped, the default state of any region.
2) Reserved - Owned by the runtime, but accessing it would cause a fault.
              Does not count against the process' memory footprint.
3) Prepared - Reserved, intended not to be backed by physical memory (though
              an OS may implement this lazily). Can transition efficiently to
              Ready. Accessing memory in such a region is undefined (may
              fault, may give back unexpected zeroes, etc.).
4) Ready - may be accessed safely.

sysAlloc:   None -> Ready
sysFree:    * -> None
sysReserve: None -> Reserved
sysMap:     Reserved -> Prepared
sysUsed:    Prepared -> Ready
sysUnused:  Ready -> Prepared
sysFault:   Ready/Prepared -> Reserved (only runtime debugging)
```

```go
/**************** [mem_linux.go] ****************/

/*
 * mmap:
 * PROT_READ - 可读
 * PROT_WRITE - 可写
 * MAP_ANON - 非文件映射，fd 可忽略（或设置为 -1），offset 必须为 0
 * MAP_PRIVATE - 私有空间，不与其他进程共享（常用于内存分配）
*/
func sysAlloc(n uintptr, sysStat *sysMemStat) unsafe.Pointer {
	p, err := mmap(nil, n, _PROT_READ|_PROT_WRITE, _MAP_ANON|_MAP_PRIVATE, -1, 0)
	if err != 0 {
		if err == _EACCES {
			print("runtime: mmap: access denied\n")
			exit(2)
		}
		if err == _EAGAIN {
			print("runtime: mmap: too much locked memory (check 'ulimit -l').\n")
			exit(2)
		}
		return nil
	}
	sysStat.add(int64(n))
	return p
}

/* 直接调用 munmap */
func sysFree(v unsafe.Pointer, n uintptr, sysStat *sysMemStat) {
	sysStat.add(-int64(n))
	munmap(v, n)
}

/* 
 * mmap:
 * PROT_NONE - 不可访问
*/
func sysReserve(v unsafe.Pointer, n uintptr) unsafe.Pointer {
	p, err := mmap(v, n, _PROT_NONE, _MAP_ANON|_MAP_PRIVATE, -1, 0)
	if err != 0 {
		return nil
	}
	return p
}

/* 
 * mmap:
 * MAP_FIXED - 传入地址不作为提示（hint），而是必须指定为该地址，如果地址不可用则失败
*/
func sysMap(v unsafe.Pointer, n uintptr, sysStat *sysMemStat) {
	sysStat.add(int64(n))

	p, err := mmap(v, n, _PROT_READ|_PROT_WRITE, _MAP_ANON|_MAP_FIXED|_MAP_PRIVATE, -1, 0)
	if err == _ENOMEM {
		throw("runtime: out of memory")
	}
	if p != v || err != 0 {
		throw("runtime: cannot map pages in arena address space")
	}
}

func sysUsed(v unsafe.Pointer, n uintptr) {
	// Partially undo the NOHUGEPAGE marks from sysUnused
	// for whole huge pages between v and v+n. This may
	// leave huge pages off at the end points v and v+n
	// even though allocations may cover these entire huge
	// pages. We could detect this and undo NOHUGEPAGE on
	// the end points as well, but it's probably not worth
	// the cost because when neighboring allocations are
	// freed sysUnused will just set NOHUGEPAGE again.
	sysHugePage(v, n)
}

/* 
 * madvise:
 * MADV_HUGEPAGE - 在给定范围内开启透明大页（THP），主要用于使用大块内存的场景
*/
func sysHugePage(v unsafe.Pointer, n uintptr) {
	if physHugePageSize != 0 {
		// Round v up to a huge page boundary.
		beg := alignUp(uintptr(v), physHugePageSize)
		// Round v+n down to a huge page boundary.
		end := alignDown(uintptr(v)+n, physHugePageSize)

		if beg < end {
			madvise(unsafe.Pointer(beg), end-beg, _MADV_HUGEPAGE)
		}
	}
}

/*
 * madvise:
 * MADV_NOHUGEPAGE - 取消透明大页
 * MADV_DONTNEED - 在将来不再访问该空间
*/
func sysUnused(v unsafe.Pointer, n uintptr) {
	// By default, Linux's "transparent huge page" support will
	// merge pages into a huge page if there's even a single
	// present regular page, undoing the effects of madvise(adviseUnused)
	// below. On amd64, that means khugepaged can turn a single
	// 4KB page to 2MB, bloating the process's RSS by as much as
	// 512X. (See issue #8832 and Linux kernel bug
	// https://bugzilla.kernel.org/show_bug.cgi?id=93111)
	//
	// To work around this, we explicitly disable transparent huge
	// pages when we release pages of the heap. However, we have
	// to do this carefully because changing this flag tends to
	// split the VMA (memory mapping) containing v in to three
	// VMAs in order to track the different values of the
	// MADV_NOHUGEPAGE flag in the different regions. There's a
	// default limit of 65530 VMAs per address space (sysctl
	// vm.max_map_count), so we must be careful not to create too
	// many VMAs (see issue #12233).
	//
	// Since huge pages are huge, there's little use in adjusting
	// the MADV_NOHUGEPAGE flag on a fine granularity, so we avoid
	// exploding the number of VMAs by only adjusting the
	// MADV_NOHUGEPAGE flag on a large granularity. This still
	// gets most of the benefit of huge pages while keeping the
	// number of VMAs under control. With hugePageSize = 2MB, even
	// a pessimal heap can reach 128GB before running out of VMAs.
	if physHugePageSize != 0 {
		// If it's a large allocation, we want to leave huge
		// pages enabled. Hence, we only adjust the huge page
		// flag on the huge pages containing v and v+n-1, and
		// only if those aren't aligned.
		var head, tail uintptr
		if uintptr(v)&(physHugePageSize-1) != 0 {
			// Compute huge page containing v.
			head = alignDown(uintptr(v), physHugePageSize)
		}
		if (uintptr(v)+n)&(physHugePageSize-1) != 0 {
			// Compute huge page containing v+n-1.
			tail = alignDown(uintptr(v)+n-1, physHugePageSize)
		}

		// Note that madvise will return EINVAL if the flag is
		// already set, which is quite likely. We ignore
		// errors.
		if head != 0 && head+physHugePageSize == tail {
			// head and tail are different but adjacent,
			// so do this in one call.
			madvise(unsafe.Pointer(head), 2*physHugePageSize, _MADV_NOHUGEPAGE)
		} else {
			// Advise the huge pages containing v and v+n-1.
			if head != 0 {
				madvise(unsafe.Pointer(head), physHugePageSize, _MADV_NOHUGEPAGE)
			}
			if tail != 0 && tail != head {
				madvise(unsafe.Pointer(tail), physHugePageSize, _MADV_NOHUGEPAGE)
			}
		}
	}

	if uintptr(v)&(physPageSize-1) != 0 || n&(physPageSize-1) != 0 {
		// madvise will round this to any physical page
		// *covered* by this range, so an unaligned madvise
		// will release more memory than intended.
		throw("unaligned sysUnused")
	}

	var advise uint32
	if debug.madvdontneed != 0 {
		advise = _MADV_DONTNEED
	} else {
		advise = atomic.Load(&adviseUnused)
	}
	if errno := madvise(v, n, int32(advise)); advise == _MADV_FREE && errno != 0 {
		// MADV_FREE was added in Linux 4.5. Fall back to MADV_DONTNEED if it is
		// not supported.
		atomic.Store(&adviseUnused, _MADV_DONTNEED)
		madvise(v, n, _MADV_DONTNEED)
	}
}

func sysFault(v unsafe.Pointer, n uintptr) {
	mmap(v, n, _PROT_NONE, _MAP_ANON|_MAP_PRIVATE|_MAP_FIXED, -1, 0)
}

```

#### 5.2 内存回收（GC）



### 6. 抢占

