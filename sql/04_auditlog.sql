-- sql/04_auditlog.sql
-- Implementa la tabla AuditLog particionada y append-only (EP-004B).

BEGIN;

-- 1. Crear tabla AuditLog particionada por fecha (rango)
CREATE TABLE IF NOT EXISTS audit_log (
    id UUID NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    provider_id VARCHAR(50) NOT NULL,
    model VARCHAR(100) NOT NULL,
    status_code INT NOT NULL,
    -- El payload será guardado cifrado vía KMS (Client-Side Encryption)
    payload_encrypted BYTEA NOT NULL,
    dek_id VARCHAR(100) NOT NULL, -- Referencia a la Data Encryption Key usada
    PRIMARY KEY (id, timestamp)
) PARTITION BY RANGE (timestamp);

-- 2. Crear trigger para forzar Append-Only (bloquear UPDATE y DELETE)
CREATE OR REPLACE FUNCTION trg_auditlog_append_only()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'audit_log is append-only: UPDATE and DELETE are strictly forbidden for compliance (EP-004B)';
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_auditlog_no_update_delete
BEFORE UPDATE OR DELETE ON audit_log
FOR EACH ROW
EXECUTE FUNCTION trg_auditlog_append_only();

-- Nota: Las particiones mensuales (ej. audit_log_2026_07) 
-- deberán ser creadas dinámicamente o por cron job.

COMMIT;
