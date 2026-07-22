# Quota Manager Specification

## 1. Responsibilities
El Quota Manager es responsable de contabilizar y autorizar el consumo de tokens y peticiones para cada proveedor. Protege las credenciales de incurrir en sobrecostos o de superar sus cuotas asignadas (diarias, mensuales, etc.).

## 2. Interface
```go
package quota

type Consumption struct {
    Tokens   int
    Requests int
}

type Manager interface {
    // Reserve verifica de forma atómica si hay cuota disponible. 
    // Retorna false si excede la cuota para la ventana temporal actual.
    Reserve(providerID string, estimate Consumption) bool
    
    // Commit confirma el consumo real post-ejecución y ajusta saldos.
    Commit(providerID string, actual Consumption) error
}
```

## 3. Comportamientos y Reglas
1. **Validación Atómica (RAM)**: Para no degradar la latencia (<50ms), las autorizaciones se realizan en memoria local.
2. **Ventana Temporal**: Los límites operan por ventanas de tiempo (ej. diario, mensual). Al expirar la ventana, la cuota se reinicia (0 consumido).
3. **Reserva (Reserve)**: Anticipa el gasto (con un estimado de tokens de entrada+salida). Si falla, la request es rechazada/ruteada a otro proveedor (HU-006 AC2 y AC4).
4. **Descuento final (Commit)**: Registra el conteo real tras completarse el stream o la respuesta síncrona. Si el real excede el estimado, la cuota de la siguiente ventana puede arrastrar déficit o bloquear futuras requests (HU-006 AC5).
5. **Persistencia**: La memoria sincroniza asincrónicamente mediante el WAL persistente (EP-009) para no perder el estado de cuota tras reinicios.
