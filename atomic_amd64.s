#include "go_asm.h"
#include "funcdata.h"
#include "textflag.h"

TEXT ·io_uring_smp_mb_fallback(SB), NOSPLIT, $0
    LOCK
    ORQ $0, 0(SP)
    RET

TEXT ·io_uring_smp_mb_mfence(SB), NOSPLIT, $0
    MFENCE
    RET
