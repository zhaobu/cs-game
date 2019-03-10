## mahjong-game-slb-monitor

* 监控所有game服务器前台的SLB服务器的健康情况
* 连续丢5个包，认为SLB异常
* 连续成功5个包，认为SLB恢复工作
* 每10秒检测一次SLB池的更新
* 并发检测功能未实现，暂时是串行的，所以未必可以在5秒内发现异常，这个时间，取决于异常的slb个数
* 服务器异常和恢复服务都会有邮件通知

* 特别注意
* 在linux系统中运行时，可能会提示“Error listening for ICMP packets: socket: permission denied”错误，解决方案：
    * 方案1: 

        ~~~
            sudo sysctl -w net.ipv4.ping_group_range="0   2147483647"
        ~~~
    * 方案2: 

        ~~~
            pinger.SetPrivileged(true)
            setcap cap_net_raw=+ep /bin/goping-binary
        ~~~