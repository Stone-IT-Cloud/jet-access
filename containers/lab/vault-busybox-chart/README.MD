# vault-busybox Helm Chart

Helm chart para desplegar HashiCorp Vault en modo producción y 2 instancias de BusyBox.

Este chart utiliza el [chart oficial de HashiCorp Vault](https://github.com/hashicorp/vault-helm) como subchart para desplegar Vault.

## Configuración

Las principales opciones de configuración se encuentran en `values.yaml`:

- `global.namespace`: Namespace donde se desplegarán los recursos.
- `vault.*`: Configuración para el subchart de Vault. Consulta la [documentación del chart oficial de Vault](https://developer.hashicorp.com/vault/docs/platform/helm) para ver todas las opciones disponibles.
- `busybox.*`: Configuración para las instancias de BusyBox.

## Uso

1.  Asegúrate de tener Helm instalado.
2.  Actualiza las dependencias del chart:
    ```bash
    helm dependency update ./containers/lab/vault-busybox-chart
    ```
3.  Instala el chart:
    ```bash
    helm install <release-name> ./containers/lab/vault-busybox-chart -n <namespace> --create-namespace
    ```
