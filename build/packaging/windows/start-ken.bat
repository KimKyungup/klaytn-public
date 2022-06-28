@echo off

set HOME=%~dp0
set CONF=%HOME%\conf

call %CONF%\ken-conf.cmd

REM Check if exist data directory
set "NOT_INIT="
IF NOT EXIST %KLAY_HOME% (
    set NOT_INIT=1
)
IF NOT EXIST %DATA_DIR% (
    set NOT_INIT=1
)

IF DEFINED NOT_INIT (
    echo "[ERROR] : ken is not initiated, Initiate ken with genesis file first."
    GOTO end
)

set OPTIONS=--networkid %NETWORK_ID%

IF DEFINED DATA_DIR (
    set OPTIONS=%OPTIONS% --datadir %DATA_DIR%
)

IF DEFINED PORT (
    set OPTIONS=%OPTIONS% --port %PORT%
)

IF DEFINED SERVER_TYPE (
    set OPTIONS=%OPTIONS% --srvtype %SERVER_TYPE%
)

IF DEFINED VERBOSITY (
    set OPTIONS=%OPTIONS% --verbosity %VERBOSITY%
)

IF DEFINED TXPOOL_EXEC_SLOTS_ALL (
    set OPTIONS=%OPTIONS% --txpool.exec-slots.all %TXPOOL_EXEC_SLOTS_ALL%
)

IF DEFINED TXPOOL_NONEXEC_SLOTS_ALL (
    set OPTIONS=%OPTIONS% --txpool.nonexec-slots.all %TXPOOL_NONEXEC_SLOTS_ALL%
)

IF DEFINED TXPOOL_EXEC_SLOTS_ACCOUNT (
    set OPTIONS=%OPTIONS% --txpool.exec-slots.account %TXPOOL_EXEC_SLOTS_ACCOUNT%
)

IF DEFINED TXPOOL_NONEXEC_SLOTS_ACCOUNT (
    set OPTIONS=%OPTIONS% --txpool.nonexec-slots.account %TXPOOL_NONEXEC_SLOTS_ACCOUNT%
)

IF DEFINED TXPOOL_LIFE_TIME (
    set OPTIONS=%OPTIONS% --txpool.lifetime %TXPOOL_LIFE_TIME%
)

IF DEFINED SYNCMODE (
    set OPTIONS=%OPTIONS% --syncmode %SYNCMODE%
)

IF DEFINED MAXCONNECTIONS (
    set OPTIONS=%OPTIONS% --maxconnections %MAXCONNECTIONS%
)

IF DEFINED LDBCACHESIZE (
    set OPTIONS=%OPTIONS% --db.leveldb.cache-size %LDBCACHESIZE%
)

IF DEFINED RPC_ENABLE (
    IF %RPC_ENABLE%==1 (
        set OPTIONS=%OPTIONS% --rpc --rpcapi %RPC_API% --rpcport %RPC_PORT% --rpcaddr %RPC_ADDR% --rpccorsdomain ^
%RPC_CORSDOMAIN% --rpcvhosts %RPC_VHOSTS%
        IF DEFINED RPC_CONCURRENCY_LIMIT (
            set OPTIONS=%OPTIONS% --rpc.concurrencylimit %RPC_CONCURRENCY_LIMIT%
        )
        IF DEFINED RPC_READ_TIMEOUT (
            set OPTIONS=%OPTIONS% --rpcreadtimeout %RPC_READ_TIMEOUT%
        )
        IF DEFINED RPC_WRITE_TIMEOUT (
            set OPTIONS=%OPTIONS% --rpcwritetimeout %RPC_WRITE_TIMEOUT%
        )
        IF DEFINED RPC_IDLE_TIMEOUT (
            set OPTIONS=%OPTIONS% --rpcidletimeout %RPC_IDLE_TIMEOUT%
        )
        IF DEFINED RPC_EXECUTION_TIMEOUT (
            set OPTIONS=%OPTIONS% --rpcexecutiontimeout %RPC_EXECUTION_TIMEOUT%
        )
    )
)

IF DEFINED WS_ENABLE (
    IF %WS_ENABLE%==1 (
        set OPTIONS=%OPTIONS% --ws --wsapi %WS_API% --wsaddr %WS_ADDR% --wsport %WS_PORT% --wsorigins %WS_ORIGINS%
    )
)

IF DEFINED METRICS (
    IF %METRICS%==1 (
        set OPTIONS=%OPTIONS% --metrics
    )
)

IF DEFINED PROMETHEUS (
    IF %PROMETHEUS%==1 (
        set OPTIONS=%OPTIONS% --prometheus
    )
)

IF DEFINED NO_DISCOVER (
    IF %NO_DISCOVER%==1 (
        set OPTIONS=%OPTIONS% --nodiscover
    )
)

IF DEFINED DB_NO_PARALLEL_WRITE (
    IF %DB_NO_PARALLEL_WRITE%==1 (
        set OPTIONS=%OPTIONS% --db.no-parallel-write
    )
)

IF DEFINED MULTICHANNEL (
    IF %MULTICHANNEL%==1 (
        set OPTIONS=%OPTIONS% --multichannel
    )
)

IF DEFINED SC_BRIDGE (
    IF %SC_BRIDGE%==1 (
        set OPTIONS=%OPTIONS% --bridge --mainbridge --bridgeport %SC_BRIDGE_PORT%
        if %SC_INDEXING%==1 (
            set OPTIONS=%OPTIONS% --childchainindexing
        )
    )
)

IF DEFINED ADDITIONAL (
    IF NOT %ADDITIONAL%=="" (
        set OPTIONS=%OPTIONS% %ADDITIONAL%
    )
)

%HOME%\bin\ken.exe %OPTIONS%

:end
@pause
