#!/usr/bin/env bash

#
# Prerequisite : install tool jq
# This script assumes that KubeDB & Component operators are installed
#
# End to end scenario to be executed on minikube, minishift or k8s cluster
# Example: ./scripts/end-to-end.sh <CLUSTER_IP> <NAMESPACE> <DEPLOYMENT_MODE>
# where CLUSTER_IP represents the external IP address exposed top of the VM
#
CLUSTER_IP=${1:-192.168.99.50}
NS=${2:-test}
MODE=${3:-dev}

SLEEP_TIME=30s
TIME=$(date +"%Y-%m-%d_%H-%M")
REPORT_FILE="result_${TIME}.txt"
EXPECTED_RESPONSE='{"status":"UP"}'
INGRESS_RESOURCES=$(kubectl get ing 2>&1)

# Test if we run on plain k8s or openshift
res=$(kubectl api-versions | grep user.openshift.io/v1)
if [ "$res" == "" ]; then
  isOpenShift="false"
else
  isOpenShift="true"
fi

if [ "$MODE" == "build" ]; then
    COMPONENT_FRUIT_BACKEND_NAME="fruit-backend-sb-build"
    COMPONENT_FRUIT_CLIENT_NAME="fruit-client-sb-build"
else
    COMPONENT_FRUIT_BACKEND_NAME="fruit-backend-sb"
    COMPONENT_FRUIT_CLIENT_NAME="fruit-client-sb"
fi

function deleteResources() {
  result=$(kubectl api-resources --verbs=list --namespaced -o name)
  for i in $result[@]
  do
    kubectl delete $i --ignore-not-found=true --all -n $1
  done
}

function listAllK8sResources() {
  result=$(kubectl api-resources --verbs=list --namespaced -o name)
  for i in $result[@]
  do
    kubectl get $i --ignore-not-found=true -n $1
  done
}

function printTitle {
  r=$(typeset i=${#1} c="=" s="" ; while ((i)) ; do ((i=i-1)) ; s="$s$c" ; done ; echo  "$s" ;)
  printf "$r\n$1\n$r\n"
}

function createRegistryCertSecret() {
cat <<EOF | kubectl apply -n ${NS} -f -
apiVersion: v1
kind: Secret
metadata:
  name: registry-certificates
  namespace: test1
type: Opaque
data:
  registry.key: >-
    LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb3dJQkFBS0NBUUVBK2xLa09mMmg1ZkMyTG5Ya3RHUUFxSmh1aWttaDFiUU1tQmZMQ3JNNTlXZkRWYlR6CkdpbUxzemkwcm4ydkFtQUY2T3VDZkZ3amthTzZXYmZHZm1ZZGhSM1A3SXVmVnlTVzFJc0h1aTFveEhwYjYzYlcKTEdLKzJzZGFaNHVCZUZxbjdDLzF4dDg5MEJPQmY3QXBiU2NkMm9nQzNYYVROWTFyWDJsZUhLWG5peW5qK0gydQpOcUdGTVQvdWlMZGliVXA4NWtjY2x5ZUUvS1hjalpPZE9GeEJhMkdDV3BuUG5aMVZUVVFxNGMxaldxQStDQklrCldyQno0aGtpeGQ0YlhDb0M5NjhWeEpwUnhRQUQ4b1VQS0h6OC92aEVmUW15c2Vlc01wT3IrZmZyRjVOMTN3L3gKUjgvbTFBMGlmUktvbTkwRFl1VHNRRGsvbXpKYU5taHowZm5YcXdJREFRQUJBb0lCQUNsVjJEU1RRMWE3QnZwcApFVmtRWW1OMHVGd3hpSDNIZTRUcCtPZE5iVGF1NmJ5UFlzMWhLdVp2YUxhZm5uU2Y4cG5odWV4Yk1xeHNBdXVwCnd5ZEFLWU8veG9QakVtN0xaZlMyK0pHdnllc1g0WEhpYjc5b2x1ZDluOW9WV1UzTWVvb3Y2VC9yd1VOSTFVdUcKaFhDMjB1RXhNSGZ0aXFWL01zblFsbS9ZbllHSEY3aCtNOUUrU0tPVmU2NWxKL3VZb2JsMmM1R3FzNStHNGhKWQpCbU50TitQcHV1bVhMNWFRQ2Nla3VxeFdoY2tzZ3E3WTdlVE1uN3pja21RY3E0T0dJY0dKeVlzRGlDbGZHVWg4CkFDazVOK1M2RTRkUElZQ0xETzRmWmptaHZzU2kxS01QY3gzRjh6ZEZkZ2NtZzlaZVRVSVFkV1NETU13WGYxZ3QKZTRPN3FlRUNnWUVBK3VPdkw1RWdORmUySnl4TG9mNEs0aGhqOGwxdGZBSTlDVHJrT0JuNUJ4dlV2QkJsZE8wLworbE5zZlNnWmgzOTFoY2owbzgvTDdtM0ZpeEQxeGJoMnhNdi9vdzhXNXZCSHQrR1dzUzBSK1dDY3V6WkpCdm95ClBjNmFDMTVVbDFGN3FsR3lkOXIvbm41M1ZPU0QvcUhpUjVWL3M4Sk5yemVGSTU2bWluQ0RGbWtDZ1lFQS8yd0EKcnpOamR5Q2t6NHg0RjBlaWRCb0J0QlpJeWd5UW42RmNjUUc1OXc0VWxjUEtvaTZOVWhiSTZQc2EvbkthUERaVApGb1ppeUNDNGVqYzVienJxSkFQMGN3VnhqYUZucWVUT2p3L2UvbzYrNGlRYTkrVmw2ZkJ2QzQzR3pSSDlSVlZZCnlVM3hDN015bThOUVNOK1RtcmtuS1ZSUjZnTndZblV6YXdnUXd2TUNnWUFNV29lMnhPT2NFREdVN2padkxJNG0Kb2VMUi9VMjF6SHBxNk81eDRMMkZYeFp6aUM4bXVjUHJ0STNqLzhSNkNvbWo0OGhBQkt4YStpYSsrVC9RMDR0dApsMG5vSW9jVEtnT3VCenFmVU1QUXpyUUk5OXhTcnFFb3IvS2YycTQ1b1RhQXBYTXZPYVphakltZHNYN2FXK2hECmRCWU1xT1dnV2hDQk4zK2wwM0p3K1FLQmdRRGxYak52SVpLY2s2L3N3WlBHTkFucWdNQXUzQ1FaYlJjaWdtRGwKQ2t2WlU4ZWdoZVlkcGZnNlUwT3dGRzYxT0d6UXpXZm52bDVPb1RPSWJMY2k3NkQ3SHFJUitEMTBsaERsUEJkUgoyVXJEQmFUY3B0ZWc3VnVMck9ITFdsSEFMZnRtbTdIVGRDNlY5eUhuUm9sK0oyZ0JkV3Q1YmNMeGhvMFJuWFhECkU4Y1ppUUtCZ0NqenlqTUZxblZ3dUZqYlNCaUhFTUcyMW1YeDlremVDdHFicnZoY29WMGx3WUNTWkY1SkNEVkwKL0krMlpOL0lEcGlTWk1KN1lKSUp0RWtaQzFaaWVFT09JNDYvU0lZektCYW94cGlCYWZ5eFIrL0doNW1Kb1NNRQpiYUFUWjdNYzFXbTc4MEpoT0s1eU4zRjdDMU1MeUllalArQWR0OHQ5dUtHUkdOaGVEZC92Ci0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0t
  registry.cert: >-
    LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURpRENDQW5DZ0F3SUJBZ0lCQ2pBTkJna3Foa2lHOXcwQkFRc0ZBREFtTVNRd0lnWURWUVFEREJ0dmNHVnUKYzJocFpuUXRjMmxuYm1WeVFERTFORE15TVRBM05qUXdIaGNOTVRneE1USTJNRFUxTWpFeFdoY05NakF4TVRJMQpNRFUxTWpFeVdqQVhNUlV3RXdZRFZRUURFd3d4TnpJdU16QXVOQzR4T0Rjd2dnRWlNQTBHQ1NxR1NJYjNEUUVCCkFRVUFBNElCRHdBd2dnRUtBb0lCQVFENlVxUTUvYUhsOExZdWRlUzBaQUNvbUc2S1NhSFZ0QXlZRjhzS3N6bjEKWjhOVnRQTWFLWXV6T0xTdWZhOENZQVhvNjRKOFhDT1JvN3BadDhaK1poMkZIYy9zaTU5WEpKYlVpd2U2TFdqRQplbHZyZHRZc1lyN2F4MXBuaTRGNFdxZnNML1hHM3ozUUU0Ri9zQ2x0SngzYWlBTGRkcE0xald0ZmFWNGNwZWVMCktlUDRmYTQyb1lVeFArNkl0Mkp0U256bVJ4eVhKNFQ4cGR5Tms1MDRYRUZyWVlKYW1jK2RuVlZOUkNyaHpXTmEKb0Q0SUVpUmFzSFBpR1NMRjNodGNLZ0wzcnhYRW1sSEZBQVB5aFE4b2ZQeisrRVI5Q2JLeDU2d3lrNnY1OStzWAprM1hmRC9GSHorYlVEU0o5RXFpYjNRTmk1T3hBT1QrYk1sbzJhSFBSK2RlckFnTUJBQUdqZ2M4d2djd3dEZ1lEClZSMFBBUUgvQkFRREFnV2dNQk1HQTFVZEpRUU1NQW9HQ0NzR0FRVUZCd01CTUF3R0ExVWRFd0VCL3dRQ01BQXcKZ1pZR0ExVWRFUVNCampDQmk0SXRaRzlqYTJWeUxYSmxaMmx6ZEhKNUxXUmxabUYxYkhRdU1UazFMakl3TVM0NApOeTR4TWpZdWJtbHdMbWx2Z2h0a2IyTnJaWEl0Y21WbmFYTjBjbmt1WkdWbVlYVnNkQzV6ZG1PQ0tXUnZZMnRsCmNpMXlaV2RwYzNSeWVTNWtaV1poZFd4MExuTjJZeTVqYkhWemRHVnlMbXh2WTJGc2dnd3hOekl1TXpBdU5DNHgKT0RlSEJLd2VCTHN3RFFZSktvWklodmNOQVFFTEJRQURnZ0VCQUhCdUVhMkIrdzVsVDFvMlVNcFFSTVRMZFd2SApiblR2TzlaSG1QR253T3M0enpLdEV2Rjl0YXhKTjJyZHFZRkE0dmpTaTVkcVFiTi9lQUVPbzJYTVByaVhiaS82CjZqd0QzNmQ3M0d5THltWmppNzlWTzJhNjlrcGRQL0hpWFlmTlZGbit3SkdjV0hjcFl5eHIwOHMza1UrL3lDVmsKeTFiNE5vYlZaMENkZmxHeE51aFNxOFU1ZGJpZXUrU1E4eTZ4UTZ6RHRnZmg5N043MEF0YzFOeFJSSlBjbVdUYQo2OUR5a2Z5MFk4azArRjRvRjloQWtVc1l0T0ZQaGlnQnBHZlZJR2tsbWNWRjh6akZpYmQzRS9icGxNdWVTdkQvCmE2TFNLNmI1UVcweEorcytIR1VNZUJ1ejNjcU1vZjU2azZRdjgvL1VTZFdib1JLOFJOZUFCKzNWZStjPQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCi0tLS0tQkVHSU4gQ0VSVElGSUNBVEUtLS0tLQpNSUlDNmpDQ0FkS2dBd0lCQWdJQkFUQU5CZ2txaGtpRzl3MEJBUXNGQURBbU1TUXdJZ1lEVlFRRERCdHZjR1Z1CmMyaHBablF0YzJsbmJtVnlRREUxTkRNeU1UQTNOalF3SGhjTk1UZ3hNVEkyTURVek9USTBXaGNOTWpNeE1USTEKTURVek9USTFXakFtTVNRd0lnWURWUVFEREJ0dmNHVnVjMmhwWm5RdGMybG5ibVZ5UURFMU5ETXlNVEEzTmpRdwpnZ0VpTUEwR0NTcUdTSWIzRFFFQkFRVUFBNElCRHdBd2dnRUtBb0lCQVFDODRNY3FYZDBoUlBBWVh1RXRwVjkvCkpBNzNrMkp4d2dXTk93T2ZqSEY2Mlc1Vzd0aDB0L2xLKzg3b2E3M0JqN2ROT3NOOCtTdUtXcm9wUGtXajI2dUoKNXpRM1EwZlRnemppTWxtL1lxMDRDayswbVFNZUhkSDBzc0xiY0kwWEhXdk45TUZrQTV1RVNzeUFZVWFVbjdiQQpoOEpHaHN5dU93SjBuUG81OFUyVHlTSXJIdnkyVTQvWVpxUEdIOVNsN0NLeTcrM0VFMHdsWlVkcWVuRG9LVmx5CmNEbGxoZkhtTjRVSThSU2g5dXRJK0dFTkgxekpOS1JsQWNXMzZMZjhWSWpJN1ZpYWY2dEQxRkZoSVMxR2V0MEgKZ1BPL3FvYzkzZ0cxbFFsVEJoemtiNkVSVmNYZW1pb2hrcXZ4ek82MHd1VzhwS2g3RXB2R0tTTmR1bUZqWktVQgpBZ01CQUFHakl6QWhNQTRHQTFVZER3RUIvd1FFQXdJQ3BEQVBCZ05WSFJNQkFmOEVCVEFEQVFIL01BMEdDU3FHClNJYjNEUUVCQ3dVQUE0SUJBUUNPdmdjUXVjaTlRUGNPWjZHWjJpR3MzdlJuTDMxOUIvQWFFS3pTWkpFaWNSbTgKSy9ydDJRKzAwWmNDdEVwNlpheUY0WXFxYWtaa3pJRCt2R0RqQUJiU1BESHJ0N0hJcmE3UGdvSjRqT3oxU1VCbwptS0NpMW9BQUxUOFRZeGhkZzErbHMybUVtRXhoOEN5b05Ud1E5RlF3NzhXRVgxZGcwLzZGMkp3N0ZxMFI4bThBCmZIeWsxV2t4T0R4RE9pQ3FuVXN2bmo1SWprL3BqVGUxZFYvZjZVVUNCQ3B3cGNNVTJScTV0eURZUFJ2MWNpZGEKcGxPa2h2SldhMktVcHozcHVVTEFPaFc3U2t5c1FaY041ODc4L0JqVENSbXdyZEM2c1BPNktyMHNnQi95bjlpNQpQWFJyR0NVZjNXbzZSMjRBNUQzRXRBQ0RTaTc1bUttTWEwVFVzVVlmCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KLS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM2akNDQWRLZ0F3SUJBZ0lCQVRBTkJna3Foa2lHOXcwQkFRc0ZBREFtTVNRd0lnWURWUVFEREJ0dmNHVnUKYzJocFpuUXRjMmxuYm1WeVFERTFORE15TVRBM05qUXdIaGNOTVRneE1USTJNRFV6T1RJMFdoY05Nak14TVRJMQpNRFV6T1RJMVdqQW1NU1F3SWdZRFZRUUREQnR2Y0dWdWMyaHBablF0YzJsbmJtVnlRREUxTkRNeU1UQTNOalF3CmdnRWlNQTBHQ1NxR1NJYjNEUUVCQVFVQUE0SUJEd0F3Z2dFS0FvSUJBUUM4NE1jcVhkMGhSUEFZWHVFdHBWOS8KSkE3M2sySnh3Z1dOT3dPZmpIRjYyVzVXN3RoMHQvbEsrODdvYTczQmo3ZE5Pc044K1N1S1dyb3BQa1dqMjZ1Sgo1elEzUTBmVGd6amlNbG0vWXEwNENrKzBtUU1lSGRIMHNzTGJjSTBYSFd2TjlNRmtBNXVFU3N5QVlVYVVuN2JBCmg4Skdoc3l1T3dKMG5QbzU4VTJUeVNJckh2eTJVNC9ZWnFQR0g5U2w3Q0t5NyszRUUwd2xaVWRxZW5Eb0tWbHkKY0RsbGhmSG1ONFVJOFJTaDl1dEkrR0VOSDF6Sk5LUmxBY1czNkxmOFZJakk3VmlhZjZ0RDFGRmhJUzFHZXQwSApnUE8vcW9jOTNnRzFsUWxUQmh6a2I2RVJWY1hlbWlvaGtxdnh6TzYwd3VXOHBLaDdFcHZHS1NOZHVtRmpaS1VCCkFnTUJBQUdqSXpBaE1BNEdBMVVkRHdFQi93UUVBd0lDcERBUEJnTlZIUk1CQWY4RUJUQURBUUgvTUEwR0NTcUcKU0liM0RRRUJDd1VBQTRJQkFRQ092Z2NRdWNpOVFQY09aNkdaMmlHczN2Um5MMzE5Qi9BYUVLelNaSkVpY1JtOApLL3J0MlErMDBaY0N0RXA2WmF5RjRZcXFha1preklEK3ZHRGpBQmJTUERIcnQ3SElyYTdQZ29KNGpPejFTVUJvCm1LQ2kxb0FBTFQ4VFl4aGRnMStsczJtRW1FeGg4Q3lvTlR3UTlGUXc3OFdFWDFkZzAvNkYySnc3RnEwUjhtOEEKZkh5azFXa3hPRHhET2lDcW5Vc3ZuajVJamsvcGpUZTFkVi9mNlVVQ0JDcHdwY01VMlJxNXR5RFlQUnYxY2lkYQpwbE9raHZKV2EyS1VwejNwdVVMQU9oVzdTa3lzUVpjTjU4NzgvQmpUQ1Jtd3JkQzZzUE82S3Iwc2dCL3luOWk1ClBYUnJHQ1VmM1dvNlIyNEE1RDNFdEFDRFNpNzVtS21NYTBUVXNVWWYKLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQ==
EOF
}

function createDockerCfgSecret() {
cat <<EOF | kubectl apply -n ${NS} -f -
apiVersion: v1
kind: Secret
metadata:
  name: builder-dockercfg
type: kubernetes.io/dockercfg
data:
  .dockercfg: >-
    eyIxNzIuMzAuNC4xODc6NTAwMCI6eyJ1c2VybmFtZSI6InNlcnZpY2VhY2NvdW50IiwicGFzc3dvcmQiOiJleUpoYkdjaU9pSlNVekkxTmlJc0ltdHBaQ0k2SWlKOS5leUpwYzNNaU9pSnJkV0psY201bGRHVnpMM05sY25acFkyVmhZMk52ZFc1MElpd2lhM1ZpWlhKdVpYUmxjeTVwYnk5elpYSjJhV05sWVdOamIzVnVkQzl1WVcxbGMzQmhZMlVpT2lKaGRYSmxJaXdpYTNWaVpYSnVaWFJsY3k1cGJ5OXpaWEoyYVdObFlXTmpiM1Z1ZEM5elpXTnlaWFF1Ym1GdFpTSTZJbUoxYVd4a1pYSXRkRzlyWlc0dGNtUnpjbU1pTENKcmRXSmxjbTVsZEdWekxtbHZMM05sY25acFkyVmhZMk52ZFc1MEwzTmxjblpwWTJVdFlXTmpiM1Z1ZEM1dVlXMWxJam9pWW5WcGJHUmxjaUlzSW10MVltVnlibVYwWlhNdWFXOHZjMlZ5ZG1salpXRmpZMjkxYm5RdmMyVnlkbWxqWlMxaFkyTnZkVzUwTG5WcFpDSTZJakEyTVRObVkySm1MVFUxTlRVdE1URmxPUzA0TVRVMkxURXdOMkkwTkdJd016VTBNQ0lzSW5OMVlpSTZJbk41YzNSbGJUcHpaWEoyYVdObFlXTmpiM1Z1ZERwaGRYSmxPbUoxYVd4a1pYSWlmUS53MDg2NHhVTEJBZGhOUF9OajJ0UUpwOVRobElPdkY3cjNhSkxUWHdROE1PVVhsXzJScUZCLUIyd1RXRzVHWjE0Z3FmOFZrQ182QzRPM1p4Mk1JSGszUUF4bXlmYm5POG9KSno4and6UXJiTERfWHA3T3QtQUlPSHV3VURFOVQ4UVY2NS1zd3FXUnVIN2R0TlRxTDR0WFZINUxveno1bThOSzNWQkh2bmRDTVgwTGdXU2ZJTjlBLUJyZk9RakRVWTFBZ1VhZkctbHpFZ0k1d3pUNm9rdXZ1bXZTdkttck1mdHlGaF9LQi1MRDB3QlFEUi1rdUI1YXR5YURyXzBDT1ZIU05TSk0zMjE2dUNFVTI5ZHhIb29IYjB0Q2YzQW9SQmIyTjZQTlhFeHM3dG8zSHJFSFJmT2E5dnRDWU56eTRRdDBiQjlQekhBSnhXdlJuZHo4WHZoWVEiLCJlbWFpbCI6InNlcnZpY2VhY2NvdW50QGV4YW1wbGUub3JnIiwiYXV0aCI6ImMyVnlkbWxqWldGalkyOTFiblE2WlhsS2FHSkhZMmxQYVVwVFZYcEpNVTVwU1hOSmJYUndXa05KTmtscFNqa3VaWGxLY0dNelRXbFBhVXB5WkZkS2JHTnROV3hrUjFaNlRETk9iR051V25CWk1sWm9XVEpPZG1SWE5UQkphWGRwWVROV2FWcFlTblZhV0ZKc1kzazFjR0o1T1hwYVdFb3lZVmRPYkZsWFRtcGlNMVoxWkVNNWRWbFhNV3hqTTBKb1dUSlZhVTlwU21oa1dFcHNTV2wzYVdFelZtbGFXRXAxV2xoU2JHTjVOWEJpZVRsNldsaEtNbUZYVG14WlYwNXFZak5XZFdSRE9YcGFWMDU1V2xoUmRXSnRSblJhVTBrMlNXMUtNV0ZYZUd0YVdFbDBaRWM1Y2xwWE5IUmpiVko2WTIxTmFVeERTbkprVjBwc1kyMDFiR1JIVm5wTWJXeDJURE5PYkdOdVduQlpNbFpvV1RKT2RtUlhOVEJNTTA1c1kyNWFjRmt5VlhSWlYwNXFZak5XZFdSRE5YVlpWekZzU1dwdmFWbHVWbkJpUjFKc1kybEpjMGx0ZERGWmJWWjVZbTFXTUZwWVRYVmhWemgyWXpKV2VXUnRiR3BhVjBacVdUSTVNV0p1VVhaak1sWjVaRzFzYWxwVE1XaFpNazUyWkZjMU1FeHVWbkJhUTBrMlNXcEJNazFVVG0xWk1rcHRURlJWTVU1VVZYUk5WRVpzVDFNd05FMVVWVEpNVkVWM1RqSkpNRTVIU1hkTmVsVXdUVU5KYzBsdVRqRlphVWsyU1c1T05XTXpVbXhpVkhCNldsaEtNbUZYVG14WlYwNXFZak5XZFdSRWNHaGtXRXBzVDIxS01XRlhlR3RhV0VscFpsRXVkekE0TmpSNFZVeENRV1JvVGxCZlRtb3lkRkZLY0RsVWFHeEpUM1pHTjNJellVcE1WRmgzVVRoTlQxVlliRjh5VW5GR1FpMUNNbmRVVjBjMVIxb3hOR2R4WmpoV2EwTmZOa00wVHpOYWVESk5TVWhyTTFGQmVHMTVabUp1VHpodlNrcDZPR3AzZWxGeVlreEVYMWh3TjA5MExVRkpUMGgxZDFWRVJUbFVPRkZXTmpVdGMzZHhWMUoxU0Rka2RFNVVjVXcwZEZoV1NEVk1iM3A2TlcwNFRrc3pWa0pJZG01a1EwMVlNRXhuVjFObVNVNDVRUzFDY21aUFVXcEVWVmt4UVdkVllXWkhMV3g2UldkSk5YZDZWRFp2YTNWMmRXMTJVM1pMYlhKTlpuUjVSbWhmUzBJdFRFUXdkMEpSUkZJdGEzVkNOV0YwZVdGRWNsOHdRMDlXU0ZOT1UwcE5Nekl4Tm5WRFJWVXlPV1I0U0c5dlNHSXdkRU5tTTBGdlVrSmlNazQyVUU1WVJYaHpOM1J2TTBoeVJVaFNaazloT1haMFExbE9lbmswVVhRd1lrSTVVSHBJUVVwNFYzWlNibVI2T0ZoMmFGbFIifSwiZG9ja2VyLXJlZ2lzdHJ5LmRlZmF1bHQuc3ZjLmNsdXN0ZXIubG9jYWw6NTAwMCI6eyJ1c2VybmFtZSI6InNlcnZpY2VhY2NvdW50IiwicGFzc3dvcmQiOiJleUpoYkdjaU9pSlNVekkxTmlJc0ltdHBaQ0k2SWlKOS5leUpwYzNNaU9pSnJkV0psY201bGRHVnpMM05sY25acFkyVmhZMk52ZFc1MElpd2lhM1ZpWlhKdVpYUmxjeTVwYnk5elpYSjJhV05sWVdOamIzVnVkQzl1WVcxbGMzQmhZMlVpT2lKaGRYSmxJaXdpYTNWaVpYSnVaWFJsY3k1cGJ5OXpaWEoyYVdObFlXTmpiM1Z1ZEM5elpXTnlaWFF1Ym1GdFpTSTZJbUoxYVd4a1pYSXRkRzlyWlc0dGNtUnpjbU1pTENKcmRXSmxjbTVsZEdWekxtbHZMM05sY25acFkyVmhZMk52ZFc1MEwzTmxjblpwWTJVdFlXTmpiM1Z1ZEM1dVlXMWxJam9pWW5WcGJHUmxjaUlzSW10MVltVnlibVYwWlhNdWFXOHZjMlZ5ZG1salpXRmpZMjkxYm5RdmMyVnlkbWxqWlMxaFkyTnZkVzUwTG5WcFpDSTZJakEyTVRObVkySm1MVFUxTlRVdE1URmxPUzA0TVRVMkxURXdOMkkwTkdJd016VTBNQ0lzSW5OMVlpSTZJbk41YzNSbGJUcHpaWEoyYVdObFlXTmpiM1Z1ZERwaGRYSmxPbUoxYVd4a1pYSWlmUS53MDg2NHhVTEJBZGhOUF9OajJ0UUpwOVRobElPdkY3cjNhSkxUWHdROE1PVVhsXzJScUZCLUIyd1RXRzVHWjE0Z3FmOFZrQ182QzRPM1p4Mk1JSGszUUF4bXlmYm5POG9KSno4and6UXJiTERfWHA3T3QtQUlPSHV3VURFOVQ4UVY2NS1zd3FXUnVIN2R0TlRxTDR0WFZINUxveno1bThOSzNWQkh2bmRDTVgwTGdXU2ZJTjlBLUJyZk9RakRVWTFBZ1VhZkctbHpFZ0k1d3pUNm9rdXZ1bXZTdkttck1mdHlGaF9LQi1MRDB3QlFEUi1rdUI1YXR5YURyXzBDT1ZIU05TSk0zMjE2dUNFVTI5ZHhIb29IYjB0Q2YzQW9SQmIyTjZQTlhFeHM3dG8zSHJFSFJmT2E5dnRDWU56eTRRdDBiQjlQekhBSnhXdlJuZHo4WHZoWVEiLCJlbWFpbCI6InNlcnZpY2VhY2NvdW50QGV4YW1wbGUub3JnIiwiYXV0aCI6ImMyVnlkbWxqWldGalkyOTFiblE2WlhsS2FHSkhZMmxQYVVwVFZYcEpNVTVwU1hOSmJYUndXa05KTmtscFNqa3VaWGxLY0dNelRXbFBhVXB5WkZkS2JHTnROV3hrUjFaNlRETk9iR051V25CWk1sWm9XVEpPZG1SWE5UQkphWGRwWVROV2FWcFlTblZhV0ZKc1kzazFjR0o1T1hwYVdFb3lZVmRPYkZsWFRtcGlNMVoxWkVNNWRWbFhNV3hqTTBKb1dUSlZhVTlwU21oa1dFcHNTV2wzYVdFelZtbGFXRXAxV2xoU2JHTjVOWEJpZVRsNldsaEtNbUZYVG14WlYwNXFZak5XZFdSRE9YcGFWMDU1V2xoUmRXSnRSblJhVTBrMlNXMUtNV0ZYZUd0YVdFbDBaRWM1Y2xwWE5IUmpiVko2WTIxTmFVeERTbkprVjBwc1kyMDFiR1JIVm5wTWJXeDJURE5PYkdOdVduQlpNbFpvV1RKT2RtUlhOVEJNTTA1c1kyNWFjRmt5VlhSWlYwNXFZak5XZFdSRE5YVlpWekZzU1dwdmFWbHVWbkJpUjFKc1kybEpjMGx0ZERGWmJWWjVZbTFXTUZwWVRYVmhWemgyWXpKV2VXUnRiR3BhVjBacVdUSTVNV0p1VVhaak1sWjVaRzFzYWxwVE1XaFpNazUyWkZjMU1FeHVWbkJhUTBrMlNXcEJNazFVVG0xWk1rcHRURlJWTVU1VVZYUk5WRVpzVDFNd05FMVVWVEpNVkVWM1RqSkpNRTVIU1hkTmVsVXdUVU5KYzBsdVRqRlphVWsyU1c1T05XTXpVbXhpVkhCNldsaEtNbUZYVG14WlYwNXFZak5XZFdSRWNHaGtXRXBzVDIxS01XRlhlR3RhV0VscFpsRXVkekE0TmpSNFZVeENRV1JvVGxCZlRtb3lkRkZLY0RsVWFHeEpUM1pHTjNJellVcE1WRmgzVVRoTlQxVlliRjh5VW5GR1FpMUNNbmRVVjBjMVIxb3hOR2R4WmpoV2EwTmZOa00wVHpOYWVESk5TVWhyTTFGQmVHMTVabUp1VHpodlNrcDZPR3AzZWxGeVlreEVYMWh3TjA5MExVRkpUMGgxZDFWRVJUbFVPRkZXTmpVdGMzZHhWMUoxU0Rka2RFNVVjVXcwZEZoV1NEVk1iM3A2TlcwNFRrc3pWa0pJZG01a1EwMVlNRXhuVjFObVNVNDVRUzFDY21aUFVXcEVWVmt4UVdkVllXWkhMV3g2UldkSk5YZDZWRFp2YTNWMmRXMTJVM1pMYlhKTlpuUjVSbWhmUzBJdFRFUXdkMEpSUkZJdGEzVkNOV0YwZVdGRWNsOHdRMDlXU0ZOT1UwcE5Nekl4Tm5WRFJWVXlPV1I0U0c5dlNHSXdkRU5tTTBGdlVrSmlNazQyVUU1WVJYaHpOM1J2TTBoeVJVaFNaazloT1haMFExbE9lbmswVVhRd1lrSTVVSHBJUVVwNFYzWlNibVI2T0ZoMmFGbFIifSwiZG9ja2VyLXJlZ2lzdHJ5LmRlZmF1bHQuc3ZjOjUwMDAiOnsidXNlcm5hbWUiOiJzZXJ2aWNlYWNjb3VudCIsInBhc3N3b3JkIjoiZXlKaGJHY2lPaUpTVXpJMU5pSXNJbXRwWkNJNklpSjkuZXlKcGMzTWlPaUpyZFdKbGNtNWxkR1Z6TDNObGNuWnBZMlZoWTJOdmRXNTBJaXdpYTNWaVpYSnVaWFJsY3k1cGJ5OXpaWEoyYVdObFlXTmpiM1Z1ZEM5dVlXMWxjM0JoWTJVaU9pSmhkWEpsSWl3aWEzVmlaWEp1WlhSbGN5NXBieTl6WlhKMmFXTmxZV05qYjNWdWRDOXpaV055WlhRdWJtRnRaU0k2SW1KMWFXeGtaWEl0ZEc5clpXNHRjbVJ6Y21NaUxDSnJkV0psY201bGRHVnpMbWx2TDNObGNuWnBZMlZoWTJOdmRXNTBMM05sY25acFkyVXRZV05qYjNWdWRDNXVZVzFsSWpvaVluVnBiR1JsY2lJc0ltdDFZbVZ5Ym1WMFpYTXVhVzh2YzJWeWRtbGpaV0ZqWTI5MWJuUXZjMlZ5ZG1salpTMWhZMk52ZFc1MExuVnBaQ0k2SWpBMk1UTm1ZMkptTFRVMU5UVXRNVEZsT1MwNE1UVTJMVEV3TjJJME5HSXdNelUwTUNJc0luTjFZaUk2SW5ONWMzUmxiVHB6WlhKMmFXTmxZV05qYjNWdWREcGhkWEpsT21KMWFXeGtaWElpZlEudzA4NjR4VUxCQWRoTlBfTmoydFFKcDlUaGxJT3ZGN3IzYUpMVFh3UThNT1VYbF8yUnFGQi1CMndUV0c1R1oxNGdxZjhWa0NfNkM0TzNaeDJNSUhrM1FBeG15ZmJuTzhvSkp6OGp3elFyYkxEX1hwN090LUFJT0h1d1VERTlUOFFWNjUtc3dxV1J1SDdkdE5UcUw0dFhWSDVMb3p6NW04TkszVkJIdm5kQ01YMExnV1NmSU45QS1CcmZPUWpEVVkxQWdVYWZHLWx6RWdJNXd6VDZva3V2dW12U3ZLbXJNZnR5RmhfS0ItTEQwd0JRRFIta3VCNWF0eWFEcl8wQ09WSFNOU0pNMzIxNnVDRVUyOWR4SG9vSGIwdENmM0FvUkJiMk42UE5YRXhzN3RvM0hyRUhSZk9hOXZ0Q1lOenk0UXQwYkI5UHpIQUp4V3ZSbmR6OFh2aFlRIiwiZW1haWwiOiJzZXJ2aWNlYWNjb3VudEBleGFtcGxlLm9yZyIsImF1dGgiOiJjMlZ5ZG1salpXRmpZMjkxYm5RNlpYbEthR0pIWTJsUGFVcFRWWHBKTVU1cFNYTkpiWFJ3V2tOSk5rbHBTamt1WlhsS2NHTXpUV2xQYVVweVpGZEtiR050Tld4a1IxWjZURE5PYkdOdVduQlpNbFpvV1RKT2RtUlhOVEJKYVhkcFlUTldhVnBZU25WYVdGSnNZM2sxY0dKNU9YcGFXRW95WVZkT2JGbFhUbXBpTTFaMVpFTTVkVmxYTVd4ak0wSm9XVEpWYVU5cFNtaGtXRXBzU1dsM2FXRXpWbWxhV0VwMVdsaFNiR041TlhCaWVUbDZXbGhLTW1GWFRteFpWMDVxWWpOV2RXUkRPWHBhVjA1NVdsaFJkV0p0Um5SYVUwazJTVzFLTVdGWGVHdGFXRWwwWkVjNWNscFhOSFJqYlZKNlkyMU5hVXhEU25Ka1YwcHNZMjAxYkdSSFZucE1iV3gyVEROT2JHTnVXbkJaTWxab1dUSk9kbVJYTlRCTU0wNXNZMjVhY0ZreVZYUlpWMDVxWWpOV2RXUkROWFZaVnpGc1NXcHZhVmx1Vm5CaVIxSnNZMmxKYzBsdGRERlpiVlo1WW0xV01GcFlUWFZoVnpoMll6SldlV1J0YkdwYVYwWnFXVEk1TVdKdVVYWmpNbFo1Wkcxc2FscFRNV2haTWs1MlpGYzFNRXh1Vm5CYVEwazJTV3BCTWsxVVRtMVpNa3B0VEZSVk1VNVVWWFJOVkVac1QxTXdORTFVVlRKTVZFVjNUakpKTUU1SFNYZE5lbFV3VFVOSmMwbHVUakZaYVVrMlNXNU9OV016VW14aVZIQjZXbGhLTW1GWFRteFpWMDVxWWpOV2RXUkVjR2hrV0Vwc1QyMUtNV0ZYZUd0YVdFbHBabEV1ZHpBNE5qUjRWVXhDUVdSb1RsQmZUbW95ZEZGS2NEbFVhR3hKVDNaR04zSXpZVXBNVkZoM1VUaE5UMVZZYkY4eVVuRkdRaTFDTW5kVVYwYzFSMW94TkdkeFpqaFdhME5mTmtNMFR6TmFlREpOU1Vock0xRkJlRzE1Wm1KdVR6aHZTa3A2T0dwM2VsRnlZa3hFWDFod04wOTBMVUZKVDBoMWQxVkVSVGxVT0ZGV05qVXRjM2R4VjFKMVNEZGtkRTVVY1V3MGRGaFdTRFZNYjNwNk5XMDRUa3N6VmtKSWRtNWtRMDFZTUV4blYxTm1TVTQ1UVMxQ2NtWlBVV3BFVlZreFFXZFZZV1pITFd4NlJXZEpOWGQ2VkRadmEzVjJkVzEyVTNaTGJYSk5ablI1Um1oZlMwSXRURVF3ZDBKUlJGSXRhM1ZDTldGMGVXRkVjbDh3UTA5V1NGTk9VMHBOTXpJeE5uVkRSVlV5T1dSNFNHOXZTR0l3ZEVObU0wRnZVa0ppTWs0MlVFNVlSWGh6TjNSdk0waHlSVWhTWms5aE9YWjBRMWxPZW5rMFVYUXdZa0k1VUhwSVFVcDRWM1pTYm1SNk9GaDJhRmxSIn19
EOF
}

function createPostgresqlCapability() {
cat <<EOF | kubectl apply -n ${NS} -f -
apiVersion: "v1"
kind: "List"
items:
- apiVersion: devexp.runtime.redhat.com/v1alpha2
  kind: Capability
  metadata:
    name: postgres-db
  spec:
    category: database
    kind: postgres
    version: "10"
    parameters:
    - name: DB_USER
      value: admin
    - name: DB_PASSWORD
      value: admin
EOF
}
function createFruitBackend() {
cat <<EOF | kubectl apply -n ${NS} -f -
apiVersion: "v1"
kind: "List"
items:
- apiVersion: devexp.runtime.redhat.com/v1alpha2
  kind: Component
  metadata:
    name: fruit-backend-sb
    labels:
      app: fruit-backend-sb
  spec:
    exposeService: true
    deploymentMode: $MODE
    buildConfig:
      url: https://github.com/snowdrop/component-operator-demo.git
      ref: master
      moduleDirName: fruit-backend-sb
    runtime: spring-boot
    version: 2.1.3
    envs:
    - name: SPRING_PROFILES_ACTIVE
      value: postgresql-kubedb
- apiVersion: "devexp.runtime.redhat.com/v1alpha2"
  kind: "Link"
  metadata:
    name: "link-to-postgres-db"
  spec:
    componentName: $COMPONENT_FRUIT_BACKEND_NAME
    kind: "Secret"
    ref: "postgres-db-config"
EOF
}
function createFruitClient() {
cat <<EOF | kubectl apply -n ${NS} -f -
---
apiVersion: "v1"
kind: "List"
items:
- apiVersion: "devexp.runtime.redhat.com/v1alpha2"
  kind: "Component"
  metadata:
    labels:
      app: "fruit-client-sb"
      version: "0.0.1-SNAPSHOT"
    name: "fruit-client-sb"
  spec:
    deploymentMode: $MODE
    buildConfig:
      url: https://github.com/snowdrop/component-operator-demo.git
      ref: master
      moduleDirName: fruit-client-sb
    runtime: "spring-boot"
    version: "2.1.3.RELEASE"
    exposeService: true
- apiVersion: "devexp.runtime.redhat.com/v1alpha2"
  kind: "Link"
  metadata:
    name: "link-to-fruit-backend"
  spec:
    kind: "Env"
    componentName: $COMPONENT_FRUIT_CLIENT_NAME
    envs:
    - name: "ENDPOINT_BACKEND"
      value: "http://fruit-backend-sb:8080/api/fruits"
EOF
}

function createAll() {
  createPostgresqlCapability
  createFruitBackend
  createFruitClient
}

printTitle "Creating the namespace"
kubectl create ns ${NS}

printTitle "Add privileged SCC to the serviceaccount postgres-db and buildbot. Required for the operators Tekton and KubeDB"
if [ "$isOpenShift" == "true" ]; then
  echo "We run on Openshift. So we will apply the SCC rule"
  oc adm policy add-scc-to-user privileged system:serviceaccount:${NS}:postgres-db
  oc adm policy add-scc-to-user privileged system:serviceaccount:${NS}:build-bot
  oc adm policy add-role-to-user edit system:serviceaccount:${NS}:build-bot
else
  echo "We DON'T run on OpenShift. So need to change SCC"
fi

printTitle "Deploy the component for the fruit-backend, link and capability"
#createRegistryCertSecret
#createDockerCfgSecret
createPostgresqlCapability
createFruitBackend
echo "Sleep ${SLEEP_TIME}"
sleep ${SLEEP_TIME}

printTitle "Deploy the component for the fruit-client, link"
createFruitClient
echo "Sleep ${SLEEP_TIME}"
sleep ${SLEEP_TIME}

printTitle "Report status : ${TIME}" > ${REPORT_FILE}

printTitle "1. Status of the resources created using the CRDs : Component, Link or Capability" >> ${REPORT_FILE}
if [ "$INGRESS_RESOURCES" == "No resources found." ]; then
  for i in components links capabilities pods deployments deploymentconfigs services routes pvc postgreses secret/postgres-db-config
  do
    printTitle "$(echo $i | tr a-z A-Z)" >> ${REPORT_FILE}
    kubectl get $i -n ${NS} >> ${REPORT_FILE}
    printf "\n" >> ${REPORT_FILE}
  done
else
  for i in components links capabilities pods deployments services ingresses pvc postgreses secret/postgres-db-config
  do
    printTitle "$(echo $i | tr a-z A-Z)" >> ${REPORT_FILE}
    kubectl get $i -n ${NS} >> ${REPORT_FILE}
    printf "\n" >> ${REPORT_FILE}
  done
fi

printTitle "2. ENV injected to the fruit backend component"
printTitle "2. ENV injected to the fruit backend component" >> ${REPORT_FILE}
until kubectl get pods -n $NS -l app=$COMPONENT_FRUIT_BACKEND_NAME | grep "Running"; do sleep 5; done
kubectl exec -n ${NS} $(kubectl get pod -n ${NS} -lapp=$COMPONENT_FRUIT_BACKEND_NAME | grep "Running" | awk '{print $1}') env | grep DB >> ${REPORT_FILE}
printf "\n" >> ${REPORT_FILE}

printTitle "3. ENV var defined for the fruit client component"
printTitle "3. ENV var defined for the fruit client component" >> ${REPORT_FILE}
until kubectl get pods -n $NS -l app=$COMPONENT_FRUIT_CLIENT_NAME | grep "Running"; do sleep 5; done
for item in $(kubectl get pod -n ${NS} -lapp=$COMPONENT_FRUIT_CLIENT_NAME --output=name); do printf "Envs for %s\n" "$item" | grep --color -E '[^/]+$' && kubectl get "$item" -n ${NS} --output=json | jq -r -S '.spec.containers[0].env[] | " \(.name)=\(.value)"' 2>/dev/null; printf "\n"; done >> ${REPORT_FILE}
printf "\n" >> ${REPORT_FILE}

if [ "$MODE" == "dev" ]; then
  printTitle "Push fruit client and backend"
  ./demo/scripts/k8s_push_start.sh fruit-backend sb ${NS}
  ./demo/scripts/k8s_push_start.sh fruit-client sb ${NS}
fi

printTitle "Wait until Spring Boot actuator health replies UP for both microservices"
for i in $COMPONENT_FRUIT_BACKEND_NAME $COMPONENT_FRUIT_CLIENT_NAME
  do
    until [ "$HTTP_BODY" == "$EXPECTED_RESPONSE" ]; do
      HTTP_RESPONSE=$(kubectl exec -n $NS $(kubectl get pod -n $NS -lapp=$i | grep "Running" | awk '{print $1}') -- curl -L -w "HTTPSTATUS:%{http_code}" -s localhost:8080/actuator/health 2>&1)
      HTTP_BODY=$(echo $HTTP_RESPONSE | sed -e 's/HTTPSTATUS\:.*//g')
      echo "$i: Response is : $HTTP_BODY, expected is : $EXPECTED_RESPONSE"
      sleep 10s
    done
done

printTitle "Curl Fruit service"
printTitle "4. Curl Fruit Endpoint service"  >> ${REPORT_FILE}

if [ "$INGRESS_RESOURCES" == "No resources found." ]; then
    echo "No ingress resources found. We run on OpenShift" >> ${REPORT_FILE}
    FRONTEND_ROUTE_URL=$(kubectl get route/fruit-client-sb -o jsonpath='{.spec.host}' -n ${NS})
    curl http://$FRONTEND_ROUTE_URL/api/client >> ${REPORT_FILE}
else
    FRONTEND_ROUTE_URL=fruit-client-sb.$CLUSTER_IP.nip.io
    curl -H "Host: fruit-client-sb" ${FRONTEND_ROUTE_URL}/api/client >> ${REPORT_FILE}
fi

# printTitle "Delete the resources components, links and capabilities"
# kubectl delete components,links,capabilities --all -n ${NS}

printTitle "End-to-end scenario executed successfully"
printTitle "To delete the resources, run:  ./demo/scripts/delete_resources.sh <NAMESPACE>"
