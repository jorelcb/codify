package audit

import domain "github.com/jorelcb/codify/internal/domain/audit"

// AuditCommitMessageForTest expone auditCommitMessage para pruebas BDD que
// viven fuera del package. Permite que tests/bdd verifiquen las reglas
// determinísticas sin necesidad de inicializar git en el filesystem.
func AuditCommitMessageForTest(sha, header string) []domain.Finding {
	return auditCommitMessage(commit{SHA: sha, Header: header})
}

// IsMergeCommitForTest expone isMergeCommit para pruebas BDD. parentsCount
// es la cantidad de parents que el commit tiene; >= 2 indica merge.
func IsMergeCommitForTest(parentsCount int) bool {
	parents := make([]string, parentsCount)
	for i := range parents {
		parents[i] = "p"
	}
	return isMergeCommit(commit{Parents: parents})
}
