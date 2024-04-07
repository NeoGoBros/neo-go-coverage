To generate a similar dump you can run the following command:

```
neo-go contract inspect -i impulse/contract.go -c
```

# Disasm

I've split the dump into two sections: one section for each function.

## Function `GetNumber`

```
INDEX    OPCODE       PARAMETER
0        INITSLOT     2 local, 1 arg    <<
3        LDARG0
4        DUP
5        ISNULL
6        JMPIF        11 (5/05)
8        SIZE
9        JMP          13 (4/04)
11       DROP
12       PUSH0
13       PUSH5
14       JMPEQ        35 (21/15)
16       PUSHDATA1    496e76616c6964206b65792073697a65 ("Invalid key size")
34       THROW
35       SYSCALL      System.Storage.GetContext (9bf667ce)
40       STLOC0
41       LDLOC0
42       PUSHDATA1    00 ("\x00")
45       DUP
46       ISTYPE       Buffer (30)
48       JMPIF        52 (4/04)
50       CONVERT      Buffer (30)
52       DUP
53       ISNULL
54       JMPIFNOT     59 (5/05)
56       DROP
57       PUSHDATA1     ("")
59       LDARG0
60       CAT
61       SWAP
62       SYSCALL      System.Storage.Get (925de831)
67       STLOC1
68       LDLOC1
69       ISNULL
70       JMPIFNOT     92 (22/16)
72       PUSHDATA1    43616e6e6f7420676574206e756d626572 ("Cannot get number")
91       THROW
92       LDLOC1
93       DUP
94       ISTYPE       Integer (21)
96       JMPIF        100 (4/04)
98       CONVERT      Integer (21)
100      RET
```

## Function `PutNumber`

```
INDEX    OPCODE       PARAMETER
101      INITSLOT     1 local, 2 arg
104      LDARG0
105      DUP
106      ISNULL
107      JMPIF        112 (5/05)
109      SIZE
110      JMP          114 (4/04)
112      DROP
113      PUSH0
114      PUSH5
115      JMPEQ        136 (21/15)
117      PUSHDATA1    496e76616c6964206b65792073697a65 ("Invalid key size")
135      THROW
136      SYSCALL      System.Storage.GetContext (9bf667ce)
141      STLOC0
142      LDLOC0
143      PUSHDATA1    00 ("\x00")
146      DUP
147      ISTYPE       Buffer (30)
149      JMPIF        153 (4/04)
151      CONVERT      Buffer (30)
153      DUP
154      ISNULL
155      JMPIFNOT     160 (5/05)
157      DROP
158      PUSHDATA1     ("")
160      LDARG0
161      CAT
162      LDARG1
163      REVERSE3
164      SYSCALL      System.Storage.Put (e63f1884)
169      RET
```
