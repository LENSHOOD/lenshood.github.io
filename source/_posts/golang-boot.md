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

### 3. 栈扩缩容

### 4. 堆

### 5. 抢占

