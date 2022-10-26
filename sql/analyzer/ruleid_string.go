// Code generated by "stringer -type=RuleId -linecomment"; DO NOT EDIT.

package analyzer

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[applyDefaultSelectLimitId-0]
	_ = x[validateOffsetAndLimitId-1]
	_ = x[validateCreateTableId-2]
	_ = x[validateExprSemId-3]
	_ = x[resolveVariablesId-4]
	_ = x[resolveNamedWindowsId-5]
	_ = x[resolveSetVariablesId-6]
	_ = x[resolveViewsId-7]
	_ = x[liftCtesId-8]
	_ = x[resolveCtesId-9]
	_ = x[liftRecursiveCtesId-10]
	_ = x[resolveDatabasesId-11]
	_ = x[resolveTablesId-12]
	_ = x[loadStoredProceduresId-13]
	_ = x[validateDropTablesId-14]
	_ = x[setTargetSchemasId-15]
	_ = x[resolveCreateLikeId-16]
	_ = x[parseColumnDefaultsId-17]
	_ = x[resolveDropConstraintId-18]
	_ = x[validateDropConstraintId-19]
	_ = x[loadCheckConstraintsId-20]
	_ = x[resolveCreateSelectId-21]
	_ = x[resolveSubqueriesId-22]
	_ = x[setViewTargetSchemaId-23]
	_ = x[resolveUnionsId-24]
	_ = x[resolveDescribeQueryId-25]
	_ = x[checkUniqueTableNamesId-26]
	_ = x[resolveTableFunctionsId-27]
	_ = x[resolveDeclarationsId-28]
	_ = x[resolveColumnDefaultsId-29]
	_ = x[validateColumnDefaultsId-30]
	_ = x[validateCreateTriggerId-31]
	_ = x[validateCreateProcedureId-32]
	_ = x[loadInfoSchemaId-33]
	_ = x[validateReadOnlyDatabaseId-34]
	_ = x[validateReadOnlyTransactionId-35]
	_ = x[validateDatabaseSetId-36]
	_ = x[validatePrivilegesId-37]
	_ = x[reresolveTablesId-38]
	_ = x[setInsertColumnsId-39]
	_ = x[validateJoinComplexityId-40]
	_ = x[resolveNaturalJoinsId-41]
	_ = x[resolveOrderbyLiteralsId-42]
	_ = x[resolveFunctionsId-43]
	_ = x[flattenTableAliasesId-44]
	_ = x[pushdownSortId-45]
	_ = x[pushdownGroupbyAliasesId-46]
	_ = x[pushdownSubqueryAliasFiltersId-47]
	_ = x[qualifyColumnsId-48]
	_ = x[resolveColumnsId-49]
	_ = x[validateCheckConstraintId-50]
	_ = x[resolveBarewordSetVariablesId-51]
	_ = x[expandStarsId-52]
	_ = x[transposeRightJoinsId-53]
	_ = x[resolveHavingId-54]
	_ = x[mergeUnionSchemasId-55]
	_ = x[flattenAggregationExprsId-56]
	_ = x[reorderProjectionId-57]
	_ = x[resolveSubqueryExprsId-58]
	_ = x[finalizeSubqueryExprsId-59]
	_ = x[replaceCrossJoinsId-60]
	_ = x[moveJoinCondsToFilterId-61]
	_ = x[evalFilterId-62]
	_ = x[optimizeDistinctId-63]
	_ = x[finalizeSubqueriesId-64]
	_ = x[finalizeUnionsId-65]
	_ = x[loadTriggersId-66]
	_ = x[processTruncateId-67]
	_ = x[resolveAlterColumnId-68]
	_ = x[resolveGeneratorsId-69]
	_ = x[removeUnnecessaryConvertsId-70]
	_ = x[assignCatalogId-71]
	_ = x[pruneColumnsId-72]
	_ = x[stripTableNameInDefaultsId-73]
	_ = x[hoistSelectExistsId-74]
	_ = x[optimizeJoinsId-75]
	_ = x[pushdownFiltersId-76]
	_ = x[subqueryIndexesId-77]
	_ = x[inSubqueryIndexesId-78]
	_ = x[pruneTablesId-79]
	_ = x[setJoinScopeLenId-80]
	_ = x[eraseProjectionId-81]
	_ = x[replaceSortPkId-82]
	_ = x[insertTopNId-83]
	_ = x[cacheSubqueryResultsId-84]
	_ = x[cacheSubqueryAliasesInJoinsId-85]
	_ = x[applyHashLookupsId-86]
	_ = x[applyHashInId-87]
	_ = x[resolveInsertRowsId-88]
	_ = x[resolvePreparedInsertId-89]
	_ = x[applyTriggersId-90]
	_ = x[applyProceduresId-91]
	_ = x[assignRoutinesId-92]
	_ = x[modifyUpdateExprsForJoinId-93]
	_ = x[applyRowUpdateAccumulatorsId-94]
	_ = x[wrapWithRollbackId-95]
	_ = x[applyFKsId-96]
	_ = x[validateResolvedId-97]
	_ = x[validateOrderById-98]
	_ = x[validateGroupById-99]
	_ = x[validateSchemaSourceId-100]
	_ = x[validateIndexCreationId-101]
	_ = x[validateOperandsId-102]
	_ = x[validateCaseResultTypesId-103]
	_ = x[validateIntervalUsageId-104]
	_ = x[validateExplodeUsageId-105]
	_ = x[validateSubqueryColumnsId-106]
	_ = x[validateUnionSchemasMatchId-107]
	_ = x[validateAggregationsId-108]
	_ = x[AutocommitId-109]
	_ = x[TrackProcessId-110]
	_ = x[parallelizeId-111]
	_ = x[clearWarningsId-112]
}

const _RuleId_name = "applyDefaultSelectLimitvalidateOffsetAndLimitvalidateCreateTablevalidateExprSemresolveVariablesresolveNamedWindowsresolveSetVariablesresolveViewsliftCtesresolveCtesliftRecursiveCtesresolveDatabasesresolveTablesloadStoredProceduresvalidateDropTablessetTargetSchemasresolveCreateLikeparseColumnDefaultsresolveDropConstraintvalidateDropConstraintloadCheckConstraintsresolveCreateSelectresolveSubqueriessetViewTargetSchemaresolveUnionsresolveDescribeQuerycheckUniqueTableNamesresolveTableFunctionsresolveDeclarationsresolveColumnDefaultsvalidateColumnDefaultsvalidateCreateTriggervalidateCreateProcedureloadInfoSchemavalidateReadOnlyDatabasevalidateReadOnlyTransactionvalidateDatabaseSetvalidatePrivilegesreresolveTablessetInsertColumnsvalidateJoinComplexityresolveNaturalJoinsresolveOrderbyLiteralsresolveFunctionsflattenTableAliasespushdownSortpushdownGroupbyAliasespushdownSubqueryAliasFiltersqualifyColumnsresolveColumnsvalidateCheckConstraintresolveBarewordSetVariablesexpandStarstransposeRightJoinsresolveHavingmergeUnionSchemasflattenAggregationExprsreorderProjectionresolveSubqueryExprsfinalizeSubqueryExprsreplaceCrossJoinsmoveJoinCondsToFilterevalFilteroptimizeDistinctfinalizeSubqueriesfinalizeUnionsloadTriggersprocessTruncateresolveAlterColumnresolveGeneratorsremoveUnnecessaryConvertsassignCatalogpruneColumnsstripTableNamesFromColumnDefaultshoistSelectExistsoptimizeJoinspushdownFilterssubqueryIndexesinSubqueryIndexespruneTablessetJoinScopeLeneraseProjectionreplaceSortPkinsertTopNcacheSubqueryResultscacheSubqueryAliasesInJoinsapplyHashLookupsapplyHashInresolveInsertRowsresolvePreparedInsertapplyTriggersapplyProceduresassignRoutinesmodifyUpdateExprsForJoinapplyRowUpdateAccumulatorsrollback triggersapplyFKsvalidateResolvedvalidateOrderByvalidateGroupByvalidateSchemaSourcevalidateIndexCreationvalidateOperandsvalidateCaseResultTypesvalidateIntervalUsagevalidateExplodeUsagevalidateSubqueryColumnsvalidateUnionSchemasMatchvalidateAggregationsaddAutocommitNodetrackProcessparallelizeclearWarnings"

var _RuleId_index = [...]uint16{0, 23, 45, 64, 79, 95, 114, 133, 145, 153, 164, 181, 197, 210, 230, 248, 264, 281, 300, 321, 343, 363, 382, 399, 418, 431, 451, 472, 493, 512, 533, 555, 576, 599, 613, 637, 664, 683, 701, 716, 732, 754, 773, 795, 811, 830, 842, 864, 892, 906, 920, 943, 970, 981, 1000, 1013, 1030, 1053, 1070, 1090, 1111, 1128, 1149, 1159, 1175, 1193, 1207, 1219, 1234, 1252, 1269, 1294, 1307, 1319, 1352, 1369, 1382, 1397, 1412, 1429, 1440, 1455, 1470, 1483, 1493, 1513, 1540, 1556, 1567, 1584, 1605, 1618, 1633, 1647, 1671, 1697, 1714, 1722, 1738, 1753, 1768, 1788, 1809, 1825, 1848, 1869, 1889, 1912, 1937, 1957, 1974, 1986, 1997, 2010}

func (i RuleId) String() string {
	if i < 0 || i >= RuleId(len(_RuleId_index)-1) {
		return "RuleId(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _RuleId_name[_RuleId_index[i]:_RuleId_index[i+1]]
}
