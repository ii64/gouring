#!/bin/python3
import re
import subprocess
from io import BytesIO
from typing import Tuple, List

from sys import (
    stdout as sys_stdout,
    stderr as sys_stderr
)

# https://dave.cheney.net/2020/05/02/mid-stack-inlining-in-go

gcflags = [
    "-m=2",
]

re_cost  = re.compile(rb"with cost (\d+) as:")
re_cost2 = re.compile(rb"cost (\d+) exceeds")

def main():
    h = subprocess.Popen(
        args=["go","build","-gcflags="+" ".join(gcflags),"."],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE)
    ret: Tuple[str, str] = h.communicate()
    stdout, stderr = ret

    lines: List[bytes] = stderr.split(b"\n")
    inlined_lines: List[bytes] = [line for line in lines if b"inline" in line]
    
    can_inline: List[Tuple[bytes, int]]    = []
    cannot_inline: List[Tuple[bytes, int]] = []
    for line in inlined_lines:
        if b"can inline" in line: 
            inline_cost = int(re_cost.findall(line)[0])
            can_inline += [ (line, inline_cost) ]
        elif b"cannot inline" in line:
            cur_cost = 0
            try:
                cur_cost = int(re_cost2.findall(line)[0])
            except: pass
            cannot_inline += [ (line, cur_cost) ]
        else:
            sys_stderr.write(b"[UNK] ")
            sys_stderr.write(line)
            sys_stderr.write(b"\n")

    # sort by cost
    # can_inline = sorted(can_inline, key=lambda v: v[1])
    # cannot_inline = sorted(cannot_inline, key=lambda v: v[1])

    for item in can_inline:
        print( (str(item[1]).encode() +b"\t"+ item[0]) .decode() )    

    print("============")

    for item in cannot_inline:
        print( (str(item[1]).encode() +b"\t"+ item[0]) .decode() )    


if __name__ == "__main__":
    main()