#!/bin/bash
# SBSS service level module init script

PROGNAME=postlog-sa
PROGPATH=/usr/sbin/
PROGCONFIG=/etc/postlog-sa/postlog-sa.ini

service_running() {
    local PID=$( ps ax | grep ${PROGNAME} | grep -v grep | grep -v bash | awk '{print $1}' )
    
    if test -z "$PID"; then
        echo 0
        
        return
    fi
    
    echo $PID
}

start_service() {
    PID=$( service_running )
    if [ $PID -gt 0 ]; then
        echo "Service ${PROGNAME} already running [$PID]"
        
        return
    fi
    
    EXECUTE=${PROGPATH}${PROGNAME}
    
    if ! test -e ${EXECUTE}; then
        echo "Service not found at ${EXECUTE}"
        exit 1
    fi
    
    if ! test -e ${PROGCONFIG}; then
        echo "Configuration file not found at ${PROGCONFIG}"
        exit 1
    fi
    
    echo "Starting ${PROGNAME}..."
    nohup ${EXECUTE} -C ${PROGCONFIG} >/dev/null 2>&1 &
}

stop_service() {
    PID=$( service_running )
    if [ $PID = 0 ]; then
        echo "Service ${PROGNAME} is not running"
        
        return
    fi
    
    kill -QUIT $PID
}

case "$1" in
    start)
        start_service
    ;;
    stop)
        stop_service
    ;;
    restart)
        stop_service
        sleep 2
        start_service
    ;;
    status)
        PID=$( service_running )
        
        if [ $PID -gt 0 ]; then
            echo "Service ${PROGNAME} is running [$PID]"
        else
            echo "Service ${PROGNAME} is not running"
        fi
    ;;
    *)
        echo "Usage: /etc/init.d/${PROGNAME} {start|stop|restart|status}"
    ;;
esac

exit 0

