#!/bin/bash
PROJECT_PATH=$1
LANGUAGE=$2
PROJECT_TYPE=$3

# Crear estructura base según lenguaje
case $LANGUAGE in
  go)
    mkdir -p $PROJECT_PATH/{cmd,internal,pkg,test,deployments,scripts,docs}
    mkdir -p $PROJECT_PATH/internal/{domain,application,infrastructure,interfaces}
    # ... más directorios específicos de Go
    ;;
  javascript)
    mkdir -p $PROJECT_PATH/{src,test,dist,docs}
    # ... estructura para JS
    ;;
  python)
    mkdir -p $PROJECT_PATH/{src,tests,docs}
    # ... estructura para Python
    ;;
esac
