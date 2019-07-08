# Installation

- Deploy tekton
```bash
kc apply -f https://storage.googleapis.com/tekton-releases/previous/v0.4.0/release.yaml

or official openshift release
kc apply -f https://gist.githubusercontent.com/vdemeester/057090166c0805e8204685b44f6eeb7c/raw/b9415b08110d3d0291250f4a93fe0c9ec09703b3/release.oc.v0.4.0.yaml
```

- Grant scc and edit role for SA `buildbot`
```bash
oc adm policy add-scc-to-user privileged system:serviceaccount:test:build-bot
oc adm policy add-role-to-user edit system:serviceaccount:test:build-bot
```

- Install task and taskRun using `buildah` tool
```bash
export NS=test
kc delete -Rf demo/scripts/tekton/buildah -n $NS
kc apply -Rf demo/scripts/tekton/buildah -n $NS

export NS=test
oc new-project $NS
oc adm policy add-scc-to-user privileged system:serviceaccount:$NS:build-bot
oc adm policy add-role-to-user edit system:serviceaccount:$NS:build-bot

oc apply -n $NS -f https://raw.githubusercontent.com/snowdrop/component-operator/master/demo/scripts/tekton/buildah/sa.yml
oc apply -n $NS -f https://raw.githubusercontent.com/snowdrop/component-operator/master/demo/scripts/tekton/buildah/task.yml
oc apply -n $NS -f https://raw.githubusercontent.com/snowdrop/component-operator/master/demo/scripts/tekton/buildah/taskrun.yml

```

- Install task and taskRun using `kaniko` tool
```bash
export NS=test
kc delete -Rf demo/scripts/tekton/kaniko -n $NS
kc apply -Rf demo/scripts/tekton/kaniko -n $NS
```



