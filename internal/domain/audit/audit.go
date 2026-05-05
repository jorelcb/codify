// Package audit define el modelo de hallazgos cuando se revisa la coherencia
// entre commits del proyecto y las convenciones documentadas (Conventional
// Commits, branches protegidas, AGENTS.md).
//
// El modo determinista (rules-only) compara cada commit contra reglas
// hard-coded. El modo LLM (--with-llm) además consulta a un modelo para
// juicios subjetivos sobre alineación con AGENTS.md. Ambos comparten el
// modelo de Finding pero se distinguen por Severity y un flag Heuristic.
package audit

// Kind clasifica los tipos de hallazgos detectables por el auditor.
type Kind string

const (
	// CommitMessageInvalidType — el commit no usa un type válido de Conventional Commits.
	CommitMessageInvalidType Kind = "commit_invalid_type"

	// CommitMessageHeaderTooLong — el header del commit excede los 72 chars recomendados.
	CommitMessageHeaderTooLong Kind = "commit_header_too_long"

	// CommitMessageTrivial — el mensaje es un placeholder ("wip", "fix", "update", "test").
	CommitMessageTrivial Kind = "commit_trivial"

	// ProtectedBranchDirectCommit — un commit cayó directamente en main/master/develop sin PR.
	ProtectedBranchDirectCommit Kind = "protected_branch_direct"

	// AgentsAlignmentIssue — modo LLM: el commit no parece alinearse con AGENTS.md.
	AgentsAlignmentIssue Kind = "agents_alignment_issue"
)

// Severity clasifica la gravedad. Mismas reglas que drift detection: el
// modo --strict del audit eleva todo a "fatal" para CI bloqueante.
type Severity string

const (
	Significant Severity = "significant"
	Minor       Severity = "minor"
)

// Finding describe un solo hallazgo del audit.
type Finding struct {
	Kind      Kind
	Severity  Severity
	CommitSHA string
	Path      string // típicamente la branch o el path del archivo
	Detail    string
	Heuristic bool // true si el finding viene del modo LLM (subjetivo)
}

// Report agrupa los findings de un único run de audit.
type Report struct {
	Findings []Finding
	// CommitsAnalyzed cuenta los commits que el auditor evaluó.
	CommitsAnalyzed int
}

// HasSignificant reporta si al menos un finding es de severidad significant.
func (r Report) HasSignificant() bool {
	for _, f := range r.Findings {
		if f.Severity == Significant {
			return true
		}
	}
	return false
}

// IsClean reporta si no hay findings.
func (r Report) IsClean() bool {
	return len(r.Findings) == 0
}

// SeverityOf devuelve la severidad asociada a un Kind.
func SeverityOf(k Kind) Severity {
	switch k {
	case CommitMessageInvalidType, CommitMessageTrivial, ProtectedBranchDirectCommit:
		return Significant
	case CommitMessageHeaderTooLong, AgentsAlignmentIssue:
		return Minor
	default:
		return Minor
	}
}
