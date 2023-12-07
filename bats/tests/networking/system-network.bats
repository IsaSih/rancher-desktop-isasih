load '../helpers/load'

local_setup() {
    skip_on_unix
}

@test 'factory reset' {
    factory_reset
}

@test 'Start up Rancher Desktop' {
    start_application
}

@test 'Verify networking tunnel is false' {
    run get_setting '.experimental.virtualMachine.networkingTunnel'
    assert_success
    assert_output false
}

@test 'Check Privileged Services - with defined host IP address' {
    if [ "$RD_LOCATION" != "system" ]; then
        skip 'This test only applies for system installation'
    fi
    localhost="127.0.0.1"
    port="55042"
    container_image="strm/helloworld-http"
    ctrctl run -d --name hello-world -p $localhost:$port:80 "${container_image}"
    IP_address=($(powershell.exe -c "Get-NetIPAddress -AddressFamily IPv4 | % { echo \$_.IPAddress }" | tr -d '\r'))
    for ip in "${IP_address}"; do
        if ["${ip}" != "${localhost}" ]; then
            run try --max 9 --delay 10 powershell.exe -c "curl.exe http://$ip:$port"
            assert_failure
            assert_output --partial "Failed to connect to "${ip}" port "${port}""
        else
            run try --max 9 --delay 10 powershell.exe -c "curl.exe http://$ip:$port"
            assert_success
            assert_output --partial "HTTP Hello World"
        fi
    done
}

@test 'Enable networking tunnel' {
    rdctl set --experimental.virtual-machine.networking-tunnel=true
    run get_setting '.experimental.virtualMachine.networkingTunnel'
    assert_success
    assert_output true
}

@test 'Disable Kubernetes' {
    rdctl set --kubernetes.enabled=false
    run get_setting '.kubernetes.enabled'
    assert_success
    assert_output false
}

@test 'Check Privileged Services - with non-defined host IP address' {
   wait_for_container_engine
   ctrctl run -d -p 8801:80 --restart=always nginx
}

@test 'Reach container UI' {
   run try --max 9 --delay 10 curl.exe --show-error http://localhost:8801
   assert_success
   assert_output --partial "Welcome to nginx"
}