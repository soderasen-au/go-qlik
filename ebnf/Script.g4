grammar Script;

options { caseInsensitive = true; }

script : ( (statement|control) ';')* EOF;
//
// Expression Rules
//
expression : unaryExpression (expressionBinaryOperator unaryExpression)* ;

    expressionUnaryOperator : Plus | Minus | LogicalNot | BitNot;
    expressionBinaryOperator : calcOp | strOp | logicalBinOp | relOp | bitBinOp | shiftOp;
        calcOp : Plus | Minus | Multiplication | Division ;
        assignOp : PlusEq | MinusEq | MultEq | DivEq;
        relSimpleOp: Less | LessEq | Greater | GreatorEq | Eq  | NotEq;
        relOp : Less | LessEq | Greater | GreatorEq | Eq  | NotEq | RelFollows | RelPrecedes;
        shiftOp : RightShift | LeftShift;
        strOp : StrConcat | StrLike;
        logicalBinOp : LogicalAnd | LogicalOr | LogicalXor | LogicalIs | LogicalIsNot;
        logicalOp : logicalBinOp | LogicalNot;
        bitBinOp : BitAnd | BitOr | BitXor;
        bitOp : BitNot | bitBinOp;
        miscOp : ScopeColon | EqEq;
        operators : calcOp | assignOp | relOp | shiftOp | logicalOp | bitOp | miscOp;

    unaryExpression
        : basicFunctionCall
        | aggregationFunctionCall
        | (expressionUnaryOperator expression)
        | expressionConstant
        | expressionFieldRef
        | expressionFieldRefFixFormat
        | ( '(' expression ')' )
        ;

        expressionConstant : StringLiteral | NumberLiteral;
        expressionFieldRef : fieldLiteral | Identifier;
            fieldLiteral
                : DoubleQuotedStringLiteral
                | AccentStringLiteral
                | SquareBracketStringLiteral
                ;
        expressionFieldRefFixFormat : (Identifier ':')* (Identifier|NumberLiteral) ('T' | 'N' | 'R' | 'I' | 'U' | 'B')? ;

basicExpression : basicUnaryExpression (expressionBinaryOperator basicUnaryExpression)* ;
    basicUnaryExpression
        : basicFunctionCall
        | (expressionUnaryOperator expression)
        | expressionConstant
        | expressionFieldRef
        | expressionFieldRefFixFormat
        | ( '(' basicExpression ')' )
        ;

///
/// Control Statement Rules
///
fieldName : Identifier | StringLiteral | fieldLiteral ;
variableName : Identifier | StringLiteral | fieldName;
whenOrUnlessExpression : ( 'when'| 'unless') expression;

control
    : callControl
    | doControl
    | loopControl
    | exitScriptControl
    | exitForControl
    | forNextControl
    | forControl
    | ifControl
    | letControl
    | subControl
    | switchControl
    ;
callControl : 'call' Identifier ( callParamList)?;
    callParamList : '(' callParamListItem (',' callParamListItem )*  ')';
        callParamListItem : ( unquotedVariableNameForCall | expression );
            unquotedVariableNameForCall : ( Identifier | NumberLiteral) ;

doControl :
    'do' (( 'while' | 'until') expression)?
        statement*
    exitDoControl?
        statement*
    loopControl
    ;
    exitDoControl : 'exit' 'do' ( whenOrUnlessExpression)?;
    loopControl : 'loop' ( ( 'while' | 'until') expression)?;

exitScriptControl : 'exit' 'script' ( whenOrUnlessExpression )?;

forControl :
    'for' ( forToControl | forEachControl)
        statement*
    exitForControl?
        statement*
    forNextControl
    ;
    forToControl : variableName '=' expression   'to' expression ( forLoopStep)?;
    forEachControl : 'each' variableName   'in' forEachList;
        forEachList : ( forEachDir | forEachFile | forEachExpression | forEachFieldValue  );
            forEachFile : 'fileList' '(' expression ')' ;
            forEachDir : 'dirList' '(' expression ')' ;
            forEachFieldValue : 'fieldValueList' '(' expression ')' ;
            forEachExpression : expression (',' expression )*;
            forLoopStep : 'step' expression;
    exitForControl : 'exit'   'for' ( whenOrUnlessExpression)?;
    forNextControl : 'next' ( variableName)?;

ifControl :
    'if' expression 'then'
        statement*
    (elseIfControl statement*)*
    (elseControl statement*)?
    endIfControl
    ;
    elseIfControl : 'elseIf' expression   'then';
    elseControl : 'else';
    endIfControl : ( 'end'   'if' | 'endIf');

letControl  : ('let' variableName '=') ( expression )?;

subControl :
    'sub' Identifier ( '(' subParamName (',' subParamName)* ')')?
        statement*
    endSubControl
    ;
    endSubControl : ( 'end'   'sub' | 'endSub' );

    exitSubControl : 'exit'   'sub' ( whenOrUnlessExpression)?;
        subParamName : Identifier;

switchControl :
    'switch' expression
    (switchCaseControl statement*)*
    switchDefaultControl?
    switchEndControl
    ;
    switchCaseControl : 'case' expression (',' expression)*;
    switchDefaultControl : 'default';
    switchEndControl : ( 'end' 'switch' | 'endSwitch' );


///
/// derived statement rules
///
anyStringOrNumber : Identifier | StringLiteral | fieldLiteral | NumberLiteral;

definitionStatement  : definitionLabel    'declare' ( ( 'field' | 'fields'))?   'definition' (definitionSetupClause | definitionUsingClause );
    definitionLabel : ( definitionName   ':');
        definitionName : ( Identifier | StringLiteral | fieldLiteral | NumberLiteral);

    definitionSetupClause : ( prefixedTagList)? ( definitionParameters)? definitionFields ( definitionGroups)?;
        prefixedTagList : 'tagged' tagList;
            tagList : (anyString | parenthesesStringList);
                parenthesesStringList : '(' anyString (',' anyString)* ')';

        definitionParameters : 'parameters' definitionParameterAssignments;
            definitionParameterAssignments : definitionParameterAssignment (',' definitionParameterAssignment)*;
                definitionParameterAssignment : definitionParameterName '=' definitionParameterValue;
                    definitionParameterValue : ( anyStringOrNumber | unaryParameter );
                        unaryParameter : Minus NumberLiteral;
                    definitionParameterName : anyString;

        definitionFields : 'fields' definitionField (',' definitionField)*;
            definitionField : definitionFieldExpression   'as' definitionFieldName ( prefixedTagList )?;
                definitionFieldExpression : definitionUnaryExpression ( expressionBinaryOperator definitionFieldExpression)?;
                    definitionUnaryExpression : ( basicFunctionCall |
                                                    aggregationFunctionCall |
                                                    ( expressionUnaryOperator definitionFieldExpression) |
                                                    expressionConstant |
                                                    definitionExpressionFieldRef |
                                                    definitionParameterName |
                                                    expressionFieldRef |
                                                    '(' definitionFieldExpression ')'
                                                    );
                        definitionExpressionFieldRef : '$' NumberLiteral;
                definitionFieldName : anyStringOrNumber;

        definitionGroups : 'groups' definitionGroup (',' definitionGroup)*;
            definitionGroup : definitionGroupFieldNames   'type' definitionGroupType   'as' definitionGroupName;
                definitionGroupFieldNames : simpleStringList;
                    simpleStringList : anyString (',' anyString)*;
                definitionGroupType : ( 'drilldown' | 'collection' );
                definitionGroupName : anyString;

    definitionUsingClause : 'using' definitionName ( 'with' definitionParameterAssignments)?;

deriveFieldStatement  : 'derive' ( ( 'field' | 'fields'))? deriveFromClause deriveUsingClause;
    deriveFromClause : 'from' ( deriveExplicitClause | deriveImplicitClause | deriveFieldClause );
        deriveFieldClause : ( ( 'field' | 'fields'))? deriveFieldList;
            deriveFieldList : simpleStringList;
        deriveExplicitClause : 'explicit' ( ( 'tag' | 'tags'))? tagList;
        deriveImplicitClause : 'implicit' ( ( 'tag' | 'tags'))?;
    deriveUsingClause : 'using' definitionName;

//
// Prefix Rules
//
tableLabel : tableName ':';
    tableName : Identifier | StringLiteral | fieldLiteral | NumberLiteral ;

anyString : Identifier | StringLiteral | fieldLiteral ;

loadPrefixOrLabel : tableLabel | loadPrefix;
    loadPrefix : addPrefix | bufferPrefix | concatenatePrefix | crossTablePrefix | firstPrefix |
                 genericPrefix | hierarchyPrefix | hierarchyBelongsToPrefix | infoPrefix |
                 innerPrefix | intervalMatchPrefix | joinPrefix | keepPrefix | leftPrefix |
                 mappingPrefix | noConcatenatePrefix | outerJoinPrefix | replacePrefix | rightPrefix |
                 samplePrefix | semanticPrefix | whenOrUnlessPrefix | mergePrefix;
        addPrefix : 'add' ( 'only')?;
        bufferPrefix : 'buffer' ( '(' bufferPrefixOption( ',' bufferPrefixOption)?  ')' )?;
            bufferPrefixOption : ( bufferPrefixOptionIncremental | bufferPrefixOptionStale );
                bufferPrefixOptionStale : 'stale' ( 'after' )? NumberLiteral ( ( 'day' | 'days' | 'hour' | 'hours') )?;
                bufferPrefixOptionIncremental : ( 'incremental' | 'inc' | 'incr' );

        concatenatePrefix : 'concatenate' ('(' tableName ')')?;
        
        crossTablePrefix : 'crossTable' '(' attributeField ',' dataField (',' qualifierFields )? ')';
            attributeField : fieldName;
            dataField : fieldName;
            qualifierFields : NumberLiteral;
        
        firstPrefix : 'first' expression;
        genericPrefix : 'generic';
        
        hierarchyPrefix : 'hierarchy' '(' nodeId ',' parentId ',' nodeName
                                          (',' (parentName)?
                                          (',' (pathSource)?
                                          (',' (pathName)?
                                          (',' (pathDelimiter)?
                                          (',' (hierarchyDepth)?
                                          )? )? )? )? )?    ')';
            nodeId : fieldName;
            parentId : fieldName;
            nodeName : fieldName;
            parentName : fieldName;
            pathSource : fieldName;
            pathName : anyString;
            pathDelimiter : anyString;
            hierarchyDepth : anyString;
        
        hierarchyBelongsToPrefix : 'hierarchyBelongsTo' '(' nodeId ',' parentId ',' nodeName ',' ancestorId ',' ancestorName (',' depthDiff)? ')';
            ancestorId : anyString;
            ancestorName : anyString;
            depthDiff : anyString;
        
        infoPrefix : ( 'bundle' ('info')? | 'info' ) (bundleImageSpec)?;
            bundleImageSpec : 'image_Size' '(' imageWidth ',' imageHeight ')';
                imageWidth : NumberLiteral;
                imageHeight : NumberLiteral;
        
        joinOrKeep : ( 'join' | 'keep' );
        
        innerPrefix : 'inner' joinOrKeep ('('tableName')')?;
        
        intervalMatchPrefix : 'intervalMatch' '(' matchField (',' keyField)* ')';
            matchField : fieldName;
            keyField : fieldName;
        
        joinPrefix : 'join' ('(' tableName')')?;
        keepPrefix : 'keep' ('(' tableName')')?;
        leftPrefix : 'left' joinOrKeep ('(' tableName')')?;
        
        mappingPrefix : 'mapping';
        mergePrefix : 'merge' ( 'only')? ('(' fieldName (',' variableName)?')')?   'on' mergeKeyList;
            mergeKeyList : fieldName (',' fieldName)*;

        noConcatenatePrefix : 'noConcatenate';
        outerJoinPrefix : 'outer' 'join' ('(' tableName ')')?;
        replacePrefix : 'replace' ( 'only')?;
        rightPrefix : 'right' joinOrKeep ('(' tableName ')')?;
        
        samplePrefix : 'sample' expression;
        semanticPrefix : 'semantic';
        
        whenOrUnlessPrefix : ( 'when' | 'unless') expression;


selectPrefixOrLabel : tableLabel | selectPrefix;
    selectPrefix : addPrefix | bufferPrefix | concatenatePrefix | crossTablePrefix | firstPrefix |
                  genericPrefix | hierarchyPrefix | hierarchyBelongsToPrefix | infoPrefix |
                  innerPrefix | intervalMatchPrefix | joinPrefix | keepPrefix | leftPrefix |
                  mappingPrefix | noConcatenatePrefix | outerJoinPrefix | replacePrefix | rightPrefix |
                  samplePrefix | semanticPrefix | whenOrUnlessPrefix | mergePrefix;

defaultPrefix : whenOrUnlessPrefix | tableLabel;




///
/// Statements Rules Helpers
///

rawText : Identifier | StringLiteral | fieldLiteral | NumberLiteral | (RawChar+);
secretRawText : RawChar+;
sensitiveRawText : RawChar+;
anonymousRawText : RawChar+;

fieldNameOrStar : Identifier | StringLiteral | fieldLiteral | '*';
fieldNumber : NumberLiteral;
aliasFieldName : Identifier | StringLiteral | fieldLiteral ;
fileName : Identifier | StringLiteral | fieldLiteral ;

sensitiveAnyString : Identifier | StringLiteral | fieldLiteral;
anyStringOrAnySymbol : Identifier | StringLiteral | fieldLiteral | (RawChar+) ;
secretAnyString : Identifier | StringLiteral | fieldLiteral;
compositeFileName : compositeFileNamePart (compositeFileNamePart)*;
    compositeFileNamePart
        : Whitespace | anyStringOrNumber
        | calcOp | StrConcat | relSimpleOp | shiftOp | assignOp | ScopeColon
        | filePartSpecialChars
        ;

        filePartSpecialChars : ')' | ',' | ':' | '!' | '~';

connectPrefix : ( 'oDBC' | 'oLEDB' );

///
/// Statements Rules
///

statement
    : aliasStatement
    | autoNumberStatement
    | binaryStatement
    | commentStatement
    | customConnectStatement
    | libConnectStatement
    | bdiConnectStatement
    | bdiLiveStatement
    | connect64Statement
    | connect32Statement
    | connectStatement
    | directoryStatement
    | disConnectStatement
    | dropFieldStatement
    | dropTableStatement
    | executeStatement
    | flushLogStatement
    | forceStatement
    | inputFieldStatement
    | includeStatement
    | letStatement
    | loadStatement
    | loosenTableStatement
    | mapUsingStatement
    | nullAsValueStatement
    | nullAsNullStatement
    | qualifyStatement
    | unqualifyStatement
    | remStatement
    | renameFieldStatement
    | renameTableStatement
    | sectionStatement
    | selectStatement
    | setStatement
    | sleepStatement
    | sqlStatement
    | sqlColumnsStatement
    | sqlTablesStatement
    | sqlTypesStatement
    | qslStatement
    | starStatement
    | storeStatement
    | tagTableStatement
    | untagTableStatement
    | tagStatement
    | untagStatement
    | traceStatement
    | unmapStatement
    | searchStatement
    | definitionStatement
    | deriveFieldStatement
    ;

aliasStatement : ( defaultPrefix )* 'alias' aliasStatementRename ( ',' aliasStatementRename )* ;
    aliasStatementRename : fieldName 'as' aliasFieldName;

autoNumberStatement : ( defaultPrefix )* 'autonumber' fieldNameOrStar ( ',' fieldNameOrStar )* ( 'using' anyString )? ;

binaryStatement : ( defaultPrefix )* 'binary' rawText;

commentStatement : ( defaultPrefix )* 'comment' ( ( 'tables' | 'table' | 'fields' | 'field' ) )? ( commentUsing | commentWith );
    commentUsing : 'using' tableName;
    commentWith : tableName 'with' anyString;

customConnectStatement : ( defaultPrefix )* ( 'customConnect' | 'cUSTOM' 'connect') 'to' customConnectString;
    customConnectString : secretAnyString;

libConnectStatement : ( defaultPrefix )* 'lIB' 'connect' 'to' libConnectString;

bdiConnectStatement : ( defaultPrefix )* 'bDI' 'connect' 'to' anonymousRawText;

bdiLiveStatement : 'iMPORT' 'lIVE' anonymousRawText;


connect64Statement : ( defaultPrefix )* ( connectPrefix )? 'connect64' 'to' connectString ( accessInfo )? ;
connect32Statement : ( defaultPrefix )* ( connectPrefix )? 'connect32' 'to' connectString ( accessInfo )? ;
connectStatement : ( defaultPrefix )* ( connectPrefix )? 'connect' 'to' connectString ( accessInfo )? ;
    connectString : secretAnyString;
    libConnectString : anyString;


    accessInfo : '(' accessItem ( ',' accessItem )* ')';
        accessItem : ( accessItemAutoCommit | accessItemCodePage | accessItemMode | accessItemPassword |
                      accessItemSSO | accessItemUserID | accessItemXPassword | accessItemXUserID );
            accessItemUserID : 'userID' 'is' secretAnyString;
            accessItemXUserID : 'xUserID' 'is' secretAnyString;
            accessItemPassword : 'password' 'is' secretAnyString;
            accessItemXPassword : 'xPassword' 'is' secretAnyString;
            accessItemMode : 'mode' 'is' 'write';
            accessItemAutoCommit : 'autoCommit' ( 'on' | 'off');
            accessItemCodePage : 'codePage' 'is' ( 'off' | 'ansi' | 'unicode' | 'oem' | NumberLiteral);
            accessItemSSO : 'sSO';


directoryStatement : ( defaultPrefix )* 'directory' rawText;

disConnectStatement : ( defaultPrefix )* 'disConnect';

dropFieldStatement : ( defaultPrefix )* 'drop' ( 'field'| 'fields') fieldName ( ',' fieldName )* ( 'from' tableName ( ',' tableName )* )? ;

dropTableStatement : ( defaultPrefix )* 'drop' ( 'table'| 'tables') tableName ( ',' tableName )* ;

executeStatement : ( defaultPrefix )* 'execute' anonymousRawText;

flushLogStatement : ( defaultPrefix )*  'flushLog';

forceStatement : ( defaultPrefix )* 'force' ( 'capitalization' | 'case' ( 'upper' | 'lower' | 'mixed' ) );

inputFieldStatement : ( defaultPrefix )* 'inputField' fieldName ( ',' fieldName )* ; 


includeStatement : '$' '(' 'include' '=' RawChar+ ')';

letStatement : ( defaultPrefix )* 'let' variableName '=' ( expression )? ;

loadStatement : ( loadPrefixOrLabel )* 'load' ( 'distinct' )? fieldList ( loadSource )? ;
    fieldList : fieldListFieldOrStar ( ',' fieldListFieldOrStar )* ;
        fieldListFieldOrStar : ( '*' | field );
        field : ( fieldRef | expression ) ( 'as'? fieldRefAlias )? ;
            fieldRef : ( Identifier | fieldLiteral );
            fieldRefAlias : ( Identifier | fieldLiteral | NumberLiteral | StringLiteral );
    loadSource : loadSourceCommon | loadSourceResident | loadSourceExtension | loadFollowing;
        loadFollowing : loadClauses ;
        loadSourceCommon : ( loadAutoGenerate | loadFrom | loadFromField | loadInline ) ( fileFormatSpec )? ( loadClauses )? ;
            loadFrom : 'from' compositeFileName;
            loadInline : 'inline' inlineData;
                inlineData : ( StringLiteral | fieldLiteral );
            loadFromField : 'from_field' '(' tableName ',' fieldName ')';
            loadAutoGenerate : 'AutoGenerate' expression;
        loadSourceResident : loadResident ( loadClauses )? ( fileOrderBy )? ;
            loadResident : 'resident' residentTableLabel;
                residentTableLabel : tableName;
        loadSourceExtension : 'extension' extensionFunction '(' ( extensionParams )? ')';
            
            extensionFunction : Identifier ( '.' Identifier )? ;
            extensionParams : ( extensionScript ( ',' extensionSourceTableSpec )? | extensionSourceTableSpec );
            extensionScript : StringLiteral;
            extensionSourceTableSpec : tableName ( '{' extensionFieldList '}' )? ;
               extensionFieldList : extensionFieldOrStar ( ',' extensionFieldOrStar )* ;
                   extensionFieldOrStar : ( '*' | extensionField );
                       extensionField : ( extensionFieldRef | extensionStringFieldRef | extensionMixedFieldRef) ( 'as' fieldRefAlias )? ;
                           extensionFieldRef : fieldRef;
                           extensionStringFieldRef : 'string' '(' fieldRef ')';
                           extensionMixedFieldRef : 'mixed' '(' fieldRef ')';

    fileFormatSpec : '(' fileFormatSpecItem ( ',' fileFormatSpecItem )* ')';
        fileFormatSpecItem : ( fileType | fileCp | fileLabel | fileQuote | fileDelimiter | fileNoEof | fileTableIs |
                               fileHeaderIs | fileRecordIs | fileCommentIs | fileTabIs | fileFilter | fileInternetURLIs | fileInternetUserAgentIs );

            fileType : ( 'biff' | 'dif' | 'fix' | 'html' | 'json' | 'kml' | 'ooxml' | 'qvd' | 'qvx' | 'txt' | 'xml' | 'xmlGeneric' | 'xmlSax' | 'xmlSimple' );
            fileCp : ( fileCpName | fileCpGeneric);
                fileCpName : ( 'ansi' | 'oem' | 'mac' | 'utf8' | 'utf7' | 'symbol' | 'unicode');
                fileCpGeneric : 'codepage' 'is' NumberLiteral;
            fileLabel : ( 'embedded' | 'explicit' | 'no' ) 'labels';
            fileQuote : ( 'msq' | 'no' 'quotes' );
            fileDelimiter : 'delimiter' 'is' ( 'spaces' | Identifier | NumberLiteral | StringLiteral | ':' );
                
                

            fileNoEof : 'no' 'eof';
            fileTableIs : 'table' 'is' tableName;
            fileHeaderIs : 'header' 'is' ( fileRecordSingleLine | fileRecordBytes | fileRecordManyLines );
            fileRecordIs : 'record' 'is' ( fileRecordSingleLine | fileRecordBytes | fileRecordManyLines );
                fileRecordSingleLine : lineKeyword;
                fileRecordBytes : NumberLiteral;
                fileRecordManyLines : NumberLiteral lineKeyword;
                    lineKeyword : ( 'line' | 'lines' );

            fileCommentIs : 'comment' 'is' anyString;
            fileTabIs : 'tab' 'is' NumberLiteral;

            fileFilter : 'filters' '(' fileFilterItem ( ',' fileFilterItem )* ')' ;
                fileFilterItem : ( fileFilterItemColSplit |
                                   fileFilterItemColumnExtract |
                                   fileFilterItemExpand |
                                   fileFilterItemInterpret |
                                   fileFilterItemRemove |
                                   fileFilterItemReplace |
                                   fileFilterItemRotate |
                                   fileFilterItemHeaderRename |
                                   fileFilterItemTranspose |
                                   fileFilterItemUnwrap );

                    fileFilterItemReplace : 'replace' '(' fileFilterItemReplaceColumn ',' fileFilterItemReplaceFromDirection ',' fileFilterStringCond ')';
                        fileFilterItemReplaceColumn : NumberLiteral;
                        fileFilterItemReplaceFromDirection : ( 'top' | 'bottom' | 'right' | 'left' );
                    fileFilterItemRemove : 'remove' '(' ( 'row'| 'col') ',' ( fileFilterRowCond | fileFilterTableIndex) ')';
                    fileFilterItemHeaderRename : 'top' '(' NumberLiteral ',' anyString ')';
                    fileFilterItemInterpret : 'interpret' '(' NumberLiteral ',' anyString ',' anyString ',' NumberLiteral ')';
                    fileFilterItemUnwrap : 'unwrap' '(' ( 'row'| 'col') ',' ( fileFilterTableIndex | fileFilterRowCond) ')';
                    fileFilterItemRotate : 'rotate' '(' ( 'left'| 'right') ')';
                    fileFilterItemColumnExtract : 'colXtr' '(' NumberLiteral ',' fileFilterRowCond ',' NumberLiteral ')';
                    fileFilterItemTranspose : 'transpose' '(' ')';
                    fileFilterItemExpand : 'expand' '(' NumberLiteral ',' anyString ',' NumberLiteral ( ',' fileFilterRowCond )? ')' ;
                    fileFilterItemColSplit : 'colSplit' '(' NumberLiteral ',' fileFilterIntArray ')';

                                    fileFilterStringCond : 'strCnd' '(' ( 'null' |
                                                                          'equal' ',' anyString |
                                                                          'contain' ',' anyString |
                                                                          'start' ',' anyString |
                                                                          'end' ',' anyString |
                                                                          'length' ',' NumberLiteral |
                                                                          'shorter' ',' NumberLiteral| 
                                                                          'longer' ',' NumberLiteral | 
                                                                          'numerical') ( fileFilterStringCondOptionList )?
                                                                     ')' ;

                                        fileFilterStringCondOptionList : ',' fileFilterStringCondOption ;
                                            fileFilterStringCondOption : ( 'not' | 'case' );

                                    fileFilterTableIndex : 'pos' '(' ( 'top'| 'bottom') ',' NumberLiteral ')';
                                    fileFilterSelectPattern : 'select' '(' NumberLiteral ',' NumberLiteral ')';

                                    fileFilterRowCond : ( fileFilterRowCondCompound | fileFilterRowCondSimple );
                                        fileFilterRowCondCompound : 'rowCnd' '(' 'compound' ',' fileFilterRowCondSimple ( ',' fileFilterRowCondSimple )* ')';
                                        fileFilterRowCondSimple : ( fileFilterRowCondCellValue | fileFilterRowCondInterval | fileFilterRowCondColMatch | fileFilterRowCondEvery );
                                            fileFilterRowCondCellValue : 'rowCnd' '(' 'cellValue' ',' NumberLiteral ',' fileFilterStringCond ')';
                                            fileFilterRowCondInterval : 'rowCnd' '(' 'interval' ',' fileFilterTableIndex ',' fileFilterTableIndex ',' fileFilterSelectPattern ')';
                                            fileFilterRowCondColMatch : 'rowCnd' '(' 'colMatch' ')';
                                            fileFilterRowCondEvery : 'rowCnd' '(' 'every' ')';

                                    fileFilterIntArray : 'intArray' '(' ( NumberLiteral ( ',' NumberLiteral )* )? ')';
            fileInternetURLIs : 'uRL' 'is' sensitiveAnyString;
            fileInternetUserAgentIs : 'userAgent' 'is' anyString;

    loadClauses : ( fileWhere | fileWhile )? ( fileGroupBy )? ;
        fileWhere : 'where' basicExpression;
        fileWhile : 'while' basicExpression;
        fileGroupBy : 'group' 'by' expression ( ',' expression )* ;

    fileOrderBy : 'order' 'by' fileOrderByFieldAscDesc ( ',' fileOrderByFieldAscDesc )* ;
        fileOrderByFieldAscDesc : ( fieldName| fieldNumber) ( ( 'asc'| 'desc') )? ;

loosenTableStatement : ( defaultPrefix )* 'loosen' ( 'table'| 'tables') tableName ( ',' tableName )* ;

mapUsingStatement : ( mapUsingPrefix )* 'map' mapUsingFieldList 'using' mapUsingMapName;
    mapUsingPrefix : addPrefix | replacePrefix | whenOrUnlessPrefix;
    mapUsingFieldList : fieldNameOrStar ( ',' fieldNameOrStar )* ;
    mapUsingMapName : anyString;

nullAsValueStatement : ( defaultPrefix )* 'nullAsValue' fieldNameOrStar ( ',' fieldNameOrStar )* ;

nullAsNullStatement : ( defaultPrefix )* 'nullAsNull' fieldNameOrStar ( ',' fieldNameOrStar )* ;

qualifyStatement : ( defaultPrefix )* 'qualify' fieldNameOrStar ( ',' fieldNameOrStar )* ;
unqualifyStatement : ( defaultPrefix )* 'unqualify' fieldNameOrStar ( ',' fieldNameOrStar )* ;

remStatement : 'rem' secretRawText;

renameFieldStatement : ( defaultPrefix )* 'rename' ( 'field'| 'fields') ( 'using' tableName | renameFieldRename ( ',' renameFieldRename )* );
    renameFieldRename : fieldName 'to' fieldName;

renameTableStatement : ( defaultPrefix )* 'rename' ( 'table'| 'tables') ( 'using' tableName | renameTableRename ( ',' renameTableRename )* );
    renameTableRename : tableName 'to' tableName;


sectionStatement : 'section' sectionType;
    sectionType : ( 'access' | 'application' );

selectStatement
    : ( selectPrefixOrLabel )*
        'select'
            fieldList
        (
            loadSource
            | 'from' '(' selectStatement ')' tableName
        )?
        (fileGroupBy)?
        (fileOrderBy)?
    ;

setStatement : ( defaultPrefix )* 'set' variableName '=' ( rawText )? ;

sleepStatement : ( defaultPrefix )* 'sleep' expression;

sqlStatement : ( selectPrefixOrLabel )* 'sQL' sensitiveRawText;

sqlColumnsStatement : ( defaultPrefix )* 'sQLColumns';
sqlTablesStatement : ( defaultPrefix )* 'sQLTables';
sqlTypesStatement : ( defaultPrefix )* 'sQLTypes';

qslStatement : ( selectPrefixOrLabel )* 'qSL' anonymousRawText;

starStatement : ( defaultPrefix )* 'star' 'is' ( anyStringOrAnySymbol )? ;

storeStatement : ( defaultPrefix )* 'store' ( ( storeFieldList 'from') )? tableName 'into' compositeFileName ( storeFormatSpec )? ;
    storeFormatSpec : '(' storeFormatSpecItem ( ',' storeFormatSpecItem )* ')';
        storeFormatSpecItem : ( storeFileType | fileDelimiter );
            storeFileType : ( 'qvd' | 'txt' | 'qvx' );

    storeFieldList : storFieldListFieldOrStar ( ',' storFieldListFieldOrStar )* ;
        storFieldListFieldOrStar : ( '*' | storeField );
        storeField : storeFieldRef ( 'as' storeFieldRefAlias )? ;
            storeFieldRef : ( Identifier | fieldLiteral | StringLiteral );
            storeFieldRefAlias : ( Identifier | fieldLiteral | NumberLiteral | StringLiteral );


tagTableStatement : ( defaultPrefix )* 'tag' tableTagSecondPart;
untagTableStatement : ( defaultPrefix )* 'untag' tableTagSecondPart;
    tableTagSecondPart : 'table' tagTableWith;
        tagTableWith : tableName ( ',' tableName )* 'with' anyString ( ',' anyString )* ;


tagStatement : ( defaultPrefix )* 'tag' tagSecondPart;
untagStatement : ( defaultPrefix )* 'untag' tagSecondPart;
    tagSecondPart : ( ( 'fields' | 'field' ) )? ( tagUsingTable | tagFieldWith );
        tagUsingTable : 'using' tableName;
        tagFieldWith : fieldName ( ',' fieldName )* 'with' anyString ( ',' anyString )* ;


traceStatement : ( defaultPrefix )* 'trace' rawText;

unmapStatement : ( defaultPrefix )* 'unmap' ( '*' | fieldName ( ',' fieldName )* );

searchStatement : ( defaultPrefix )* 'search' ( 'include' | 'exclude' ) fieldNameOrStar ( ',' fieldNameOrStar )* ;





///
/// Lexer rules
///
Plus : '+';
Minus : '-';
Multiplication : '*';
Division : '/';

PlusEq : '+=';
MinusEq : '-=';
MultEq : '*=';
DivEq : '/=';

Less : '<';
LessEq : '<=';
Greater : '>';
GreatorEq : '>=';
Eq : '=';
NotEq : '<>';
RelFollows : 'follows';
RelPrecedes : 'precedes';

RightShift : '>>';
LeftShift : '<<';

StrConcat : '&';
StrLike : 'like';

LogicalIs : 'is';
LogicalIsNot: 'is' 'not';
LogicalNot : 'not';
LogicalAnd : 'and';
LogicalOr : 'or';
LogicalXor : 'xor';

BitNot : 'bitnot';
BitAnd : 'bitand';
BitOr : 'bitor';
BitXor : 'bitxor';

ScopeColon : '::';
EqEq : '==';

Whitespace
    : [ \t]+ -> channel(HIDDEN)
    ;

Newline
    : ('\r' '\n'? | '\n') -> channel(HIDDEN)
    ;

BlockComment
    : '/*' .*? '*/' -> channel(HIDDEN)
    ;

LineComment
    : '//' ~[\r\n]* -> channel(HIDDEN)
    ;

NumberLiteral
    : IntegerConstant
    | FloatingConstant
    ;

DoubleQuotedStringLiteral
    : '"' SCharSequence*? '"'
    ;

AccentStringLiteral
    : '`' CCharSequence*? '`'
    ;

SquareBracketStringLiteral
    : '[' CCharSequence*? ']'
    ;

StringLiteral
    : '\'' CCharSequence? '\''
    ;

fragment EncodingPrefix
    : 'u8'
    | 'u'
    | 'L'
    ;

fragment SCharSequence
    : SChar+
    ;

fragment SChar
    : ~["\\\r\n]
    | EscapeSequence
    | '\\\n'   // Added line
    | '\\\r\n' // Added line
    ;
fragment EscapeSequence
    : SimpleEscapeSequence
    | OctalEscapeSequence
    | HexadecimalEscapeSequence
    | UniversalCharacterName
    ;

fragment SimpleEscapeSequence
    : '\\' ['"?abfnrtv\\]
    ;

fragment OctalEscapeSequence
    : '\\' OctalDigit OctalDigit? OctalDigit?
    ;

fragment HexadecimalEscapeSequence
    : '\\x' HexadecimalDigit+
    ;

Identifier
    : IdentifierNondigit (IdentifierNondigit | Digit)*
    ;

fragment IdentifierNondigit
    : Nondigit
    | UniversalCharacterName
    //|   // other implementation-defined characters...
    ;

fragment Nondigit
    : [a-z_]
    ;

fragment Digit
    : [0-9]
    ;

fragment UniversalCharacterName
    : '\\u' HexQuad
    | '\\U' HexQuad HexQuad
    ;
fragment HexQuad
    : HexadecimalDigit HexadecimalDigit HexadecimalDigit HexadecimalDigit
    ;

fragment IntegerConstant
    : DecimalConstant IntegerSuffix?
    | OctalConstant IntegerSuffix?
    | HexadecimalConstant IntegerSuffix?
    | BinaryConstant
    ;

fragment BinaryConstant
    : '0' 'b' [0-1]+
    ;

fragment DecimalConstant
    : NonzeroDigit Digit*
    ;

fragment OctalConstant
    : '0' OctalDigit*
    ;

fragment HexadecimalConstant
    : HexadecimalPrefix HexadecimalDigit+
    ;

fragment HexadecimalPrefix
    : '0' 'x'
    ;

fragment NonzeroDigit
    : [1-9]
    ;

fragment OctalDigit
    : [0-7]
    ;

fragment HexadecimalDigit
    : [0-9a-f]
    ;

fragment IntegerSuffix
    : UnsignedSuffix LongSuffix?
    | UnsignedSuffix LongLongSuffix
    | LongSuffix UnsignedSuffix?
    | LongLongSuffix UnsignedSuffix?
    ;

fragment UnsignedSuffix
    : [u]
    ;

fragment LongSuffix
    : [l]
    ;

fragment LongLongSuffix
    : 'll'
    | 'LL'
    ;

fragment FloatingConstant
    : DecimalFloatingConstant
    | HexadecimalFloatingConstant
    ;

fragment DecimalFloatingConstant
    : FractionalConstant ExponentPart? FloatingSuffix?
    | DigitSequence ExponentPart FloatingSuffix?
    ;

fragment HexadecimalFloatingConstant
    : HexadecimalPrefix (HexadecimalFractionalConstant | HexadecimalDigitSequence) BinaryExponentPart FloatingSuffix?
    ;

fragment FractionalConstant
    : DigitSequence? '.' DigitSequence
    | DigitSequence '.'
    ;

fragment ExponentPart
    : [e] Sign? DigitSequence
    ;

fragment Sign
    : [+-]
    ;

DigitSequence
    : Digit+
    ;

fragment HexadecimalFractionalConstant
    : HexadecimalDigitSequence? '.' HexadecimalDigitSequence
    | HexadecimalDigitSequence '.'
    ;

fragment BinaryExponentPart
    : [p] Sign? DigitSequence
    ;

fragment HexadecimalDigitSequence
    : HexadecimalDigit+
    ;

fragment FloatingSuffix
    : [fl]
    ;

fragment CharacterConstant
    : '\'' CCharSequence '\''
    | 'L\'' CCharSequence '\''
    | 'u\'' CCharSequence '\''
    | 'U\'' CCharSequence '\''
    ;

fragment CCharSequence
    : CChar+
    ;

fragment CChar
    : ~['\\\r\n]
    | EscapeSequence
    ;

RawChar
    : ~[\\\r\n]
    | EscapeSequence
    ;

///
/// Function Rules
///
basicFunctionCall
    : acos_Func
	| acosh_Func
	| addMonths_Func
	| addYears_Func
	| age_Func
	| alt_Func
	| applyCodepage_Func
	| applyMap_Func
	| aRGB_Func
	| asin_Func
	| asinh_Func
	| atan_Func
	| atan2_Func
	| atanh_Func
	| attribute_Func
	| author_Func
	| autoNumber_Func
	| autoNumberHash128_Func
	| autoNumberHash256_Func
	| betaDensity_Func
	| betaDist_Func
	| betaInv_Func
	| binomDist_Func
	| binomFrequency_Func
	| binomInv_Func
	| bitCount_Func
	| black_Func
	| blackAndSchole_Func
	| blue_Func
	| brown_Func
	| capitalize_Func
	| case_Func
	| ceil_Func
	| chiDensity_Func
	| chiDist_Func
	| chiInv_Func
	| chr_Func
	| class_Func
	| clientPlatform_Func
	| coalesce_Func
	| color_Func
	| colorMapHue_Func
	| colorMapJet_Func
	| colorMix1_Func
	| colorMix2_Func
	| combin_Func
	| computerName_Func
	| connectString_Func
	| convertToLocalTime_Func
	| cos_Func
	| cosh_Func
	| cyan_Func
	| darkGray_Func
	| date_Func
	| dateH_Func
	| day_Func
	| dayEnd_Func
	| daylightSaving_Func
	| dayName_Func
	| dayNumberOfQuarter_Func
	| dayNumberOfYear_Func
	| dayStart_Func
	| div_Func
	| distinct_on_Func
	| documentName_Func
	| documentPath_Func
	| documentTitle_Func
	| dual_Func
	| e_Func
	| elapsedSeconds_Func
//	| emptyIsNull_Func
	| engineVersion_Func
	| evaluate_Func
	| even_Func
	| exists_Func
	| exp_Func
	| fAbs_Func
	| fact_Func
	| false_Func
	| fastMatch_Func
	| fDensity_Func
	| fDist_Func
	| fieldElemNo_Func
	| fieldIndex_Func
	| fieldName_Func
	| fieldNumber_Func
	| fieldValue_Func
	| fieldValueCount_Func
	| fileBaseName_Func
	| fileDir_Func
	| fileExtension_Func
	| fileName_Func
	| filePath_Func
	| fileSize_Func
	| fileTime_Func
	| findOneOf_Func
	| fInv_Func
	| firstWorkDate_Func
	| floor_Func
	| fMod_Func
	| frac_Func
	| fV_Func
	| gammaDensity_Func
	| gammaDist_Func
	| gammaInv_Func
	| geoAggrGeometry_Func
	| geoBoundingBox_Func
	| geoCountVertex_Func
	| geoGetBoundingBox_Func
	| geoGetPolygonCenter_Func
	| geoInvProjectGeometry_Func
	| geoMakePoint_Func
	| geoProject_Func
	| geoProjectGeometry_Func
	| geoReduceGeometry_Func
	| getCollationLocale_Func
	| getDataModelHash_Func
	| getFolderPath_Func
	| getObjectField_Func
	| gMT_Func
	| green_Func
	| hash128_Func
	| hash160_Func
	| hash256_Func
	| hCNoRows_Func
	| hCValue_Func
	| hour_Func
	| hSL_Func
	| if_Func
	| inDay_Func
	| inDayToTime_Func
	| index_Func
	| inLunarWeek_Func
	| inLunarWeekToDate_Func
	| inMonth_Func
	| inMonths_Func
	| inMonthsToDate_Func
	| inMonthToDate_Func
	| inQuarter_Func
	| inQuarterToDate_Func
	| interval_Func
	| intervalH_Func
	| inWeek_Func
	| inWeekToDate_Func
	| inYear_Func
	| inYearToDate_Func
	| isJson_Func
	| isNull_Func
	| isNum_Func
	| isPartialReload_Func
	| isText_Func
	| iterNo_Func
	| jsonGet_Func
	| jsonSet_Func
	| keepChar_Func
	| lastWorkDate_Func
	| left_Func
	| len_Func
	| levenshteinDist_Func
	| lightBlue_Func
	| lightCyan_Func
	| lightGray_Func
	| lightGreen_Func
	| lightMagenta_Func
	| lightRed_Func
	| localTime_Func
	| log_Func
	| log10_Func
	| lookup_Func
	| lower_Func
	| lTrim_Func
	| lunarWeekEnd_Func
	| lunarWeekName_Func
	| lunarWeekStart_Func
	| magenta_Func
	| makeDate_Func
	| makeTime_Func
	| makeWeekDate_Func
	| mapSubString_Func
	| match_Func
	| mid_Func
	| minute_Func
	| mixMatch_Func
	| mod_Func
	| money_Func
	| moneyH_Func
	| month_Func
	| monthEnd_Func
	| monthName_Func
	| monthsEnd_Func
	| monthsName_Func
	| monthsStart_Func
	| monthStart_Func
	| netWorkDays_Func
	| noOfFields_Func
	| noOfRows_Func
	| noOfTables_Func
	| normDist_Func
	| normInv_Func
	| now_Func
	| nPer_Func
	| null_Func
	| num_Func
	| numH_Func
	| numAvg_Func
	| numCount_Func
	| numMax_Func
	| numMin_Func
	| numSum_Func
	| odd_Func
	| ord_Func
	| oSUser_Func
	| peek_Func
	| permut_Func
	| pi_Func
	| pick_Func
	| pmt_Func
	| poissonDensity_Func
	| poissonDist_Func
	| poissonFrequency_Func
	| poissonInv_Func
	| pow_Func
	| previous_Func
	| productVersion_Func
	| purgeChar_Func
	| pV_Func
	| qlikTechBlue_Func
	| qlikTechGray_Func
	| qlikViewVersion_Func
	| quarterEnd_Func
	| quarterName_Func
	| quarterStart_Func
	| qvdCreateTime_Func
	| qvdFieldName_Func
	| qvdNoOfFields_Func
	| qvdNoOfRecords_Func
	| qvdTableName_Func
	| qVUser_Func
	| rand_Func
	| rangeAvg_Func
	| rangeCorrel_Func
	| rangeCount_Func
	| rangeFractile_Func
	| rangeFractileExc_Func
	| rangeIrr_Func
	| rangeKurtosis_Func
	| rangeMax_Func
	| rangeMaxString_Func
	| rangeMin_Func
	| rangeMinString_Func
	| rangeMissingCount_Func
	| rangeMode_Func
	| rangeNpv_Func
	| rangeNullCount_Func
	| rangeNumericCount_Func
	| rangeOnly_Func
	| rangeSkew_Func
	| rangeStDev_Func
	| rangeSum_Func
	| rangeTextCount_Func
	| rangeXirr_Func
	| rangeXnpv_Func
	| rate_Func
	| recNo_Func
	| red_Func
	| reloadTime_Func
	| repeat_Func
	| replace_Func
	| rGB_Func
	| right_Func
	| round_Func
	| rowNo_Func
	| rTrim_Func
	| second_Func
	| setDateYear_Func
	| setDateYearMonth_Func
	| sign_Func
	| sin_Func
	| sinh_Func
	| sqr_Func
	| sqrt_Func
	| subField_Func
	| subStringCount_Func
	| sysColor_Func
	| tableName_Func
	| tableNumber_Func
	| tan_Func
	| tanh_Func
	| tDensity_Func
	| tDist_Func
	| text_Func
	| textBetween_Func
	| time_Func
	| timeH_Func
	| timestamp_Func
	| timestampH_Func
	| timeZone_Func
	| tInv_Func
	| today_Func
	| trim_Func
	| true_Func
	| upper_Func
	| uTC_Func
	| week_Func
	| weekDay_Func
	| weekEnd_Func
	| weekName_Func
	| weekStart_Func
	| weekYear_Func
	| white_Func
	| wildMatch_Func
	| year_Func
	| year2Date_Func
	| yearEnd_Func
	| yearName_Func
	| yearStart_Func
	| yearToDate_Func
	| yellow_Func
	| otherUnknownFuncCall
	;

aggregationFunctionCall
    : avg_Func
	| chi2Test_Chi2_Func
	| chi2Test_DF_Func
	| chi2Test_p_Func
	| concat_Func
	| correl_Func
	| count_Func
	| firstSortedValue_Func
	| firstValue_Func
	| fractile_Func
	| fractileExc_Func
	| irr_Func
	| kurtosis_Func
	| lastValue_Func
	| linEst_B_Func
	| linEst_DF_Func
	| linEst_F_Func
	| linEst_M_Func
	| linEst_R2_Func
	| linEst_SEB_Func
	| linEst_SEM_Func
	| linEst_SEY_Func
	| linEst_SSReg_Func
	| linEst_SSResid_Func
	| max_Func
	| maxString_Func
	| median_Func
	| min_Func
	| minString_Func
	| missingCount_Func
	| mode_Func
	| npv_Func
	| nullCount_Func
	| numericCount_Func
	| only_Func
	| skew_Func
	| stDev_Func
	| stErr_Func
	| stEYX_Func
	| sum_Func
	| textCount_Func
	| tTest1_Conf_Func
	| tTest1_DF_Func
	| tTest1_Dif_Func
	| tTest1_Lower_Func
	| tTest1_Sig_Func
	| tTest1_StErr_Func
	| tTest1_t_Func
	| tTest1_Upper_Func
	| tTest1w_Conf_Func
	| tTest1w_DF_Func
	| tTest1w_Dif_Func
	| tTest1w_Lower_Func
	| tTest1w_Sig_Func
	| tTest1w_StErr_Func
	| tTest1w_t_Func
	| tTest1w_Upper_Func
	| tTest_Conf_Func
	| tTest_DF_Func
	| tTest_Dif_Func
	| tTest_Lower_Func
	| tTest_Sig_Func
	| tTest_StErr_Func
	| tTest_t_Func
	| tTest_Upper_Func
	| tTestw_Conf_Func
	| tTestw_DF_Func
	| tTestw_Dif_Func
	| tTestw_Lower_Func
	| tTestw_Sig_Func
	| tTestw_StErr_Func
	| tTestw_t_Func
	| tTestw_Upper_Func
	| xirr_Func
	| xnpv_Func
	| zTest_Conf_Func
	| zTest_Dif_Func
	| zTest_Lower_Func
	| zTest_Sig_Func
	| zTest_StErr_Func
	| zTest_Upper_Func
	| zTest_z_Func
	| zTestw_Conf_Func
	| zTestw_Dif_Func
	| zTestw_Lower_Func
	| zTestw_Sig_Func
	| zTestw_StErr_Func
	| zTestw_Upper_Func
	| zTestw_z_Func
    ;

parameterList : unaryExpression ( ',' unaryExpression)* ;

pi_Func :  ( 'pi' '(' ) ')';
null_Func : 'null' ('(' ')' )?;
e_Func :  ( 'e' '(' ) ')';
true_Func :  ( 'true' '(' ) ')';
false_Func :  ( 'false' '(' ) ')';
daylightSaving_Func :  ( 'daylightSaving' '(' ) ')';
timeZone_Func :  ( 'timeZone' '(' ) ')';
rand_Func :  ( 'rand' '(' ) ')';
pow_Func :  ( 'pow' '(' ) expression ',' expression  ')';
atan2_Func :  ( 'atan2' '(' ) expression ',' expression  ')';
fMod_Func :  ( 'fMod' '(' ) expression ',' expression  ')';
acos_Func :  ( 'acos' '(' ) expression  ')';
asin_Func :  ( 'asin' '(' ) expression  ')';
asinh_Func :  ( 'asinh' '(' ) expression  ')';
atan_Func :  ( 'atan' '(' ) expression  ')';
acosh_Func :  ( 'acosh' '(' ) expression  ')';
cos_Func :  ( 'cos' '(' ) expression  ')';
cosh_Func :  ( 'cosh' '(' ) expression  ')';
exp_Func :  ( 'exp' '(' ) expression  ')';
fAbs_Func :  ( 'fAbs' '(' ) expression  ')';
floor_Func :  ( 'floor' '(' ) expression (',' expression (',' expression )?)? ')';
ceil_Func :  ( 'ceil' '(' ) expression (',' expression (',' expression )?)? ')';
round_Func :  ( 'round' '(' ) expression (',' expression (',' expression )?)? ')';
pV_Func :  ( 'pV' '(' ) expression ',' expression ',' expression (',' expression (',' expression )?)? ')';
fV_Func :  ( 'fV' '(' ) expression ',' expression ',' expression (',' expression (',' expression )?)? ')';
pmt_Func :  ( 'pmt' '(' ) expression ',' expression ',' expression (',' expression (',' expression )?)? ')';
nPer_Func :  ( 'nPer' '(' ) expression ',' expression ',' expression (',' expression (',' expression )?)? ')';
rate_Func :  ( 'rate' '(' ) expression ',' expression ',' expression (',' expression (',' expression )?)? ')';
log_Func :  ( 'log' '(' ) expression  ')';
log10_Func :  ( 'log10' '(' ) expression  ')';
sin_Func :  ( 'sin' '(' ) expression  ')';
sinh_Func :  ( 'sinh' '(' ) expression  ')';
sqrt_Func :  ( 'sqrt' '(' ) expression  ')';
sqr_Func :  ( 'sqr' '(' ) expression  ')';
atanh_Func :  ( 'atanh' '(' ) expression  ')';
tan_Func :  ( 'tan' '(' ) expression  ')';
tanh_Func :  ( 'tanh' '(' ) expression  ')';
sign_Func :  ( 'sign' '(' ) expression  ')';
fact_Func :  ( 'fact' '(' ) expression  ')';
permut_Func :  ( 'permut' '(' ) expression ',' expression  ')';
combin_Func :  ( 'combin' '(' ) expression ',' expression  ')';
black_Func :  ( 'black' '(' )( expression )? ')';
blue_Func :  ( 'blue' '(' )( expression )? ')';
green_Func :  ( 'green' '(' )( expression )? ')';
cyan_Func :  ( 'cyan' '(' )( expression )? ')';
red_Func :  ( 'red' '(' )( expression )? ')';
magenta_Func :  ( 'magenta' '(' )( expression )? ')';
brown_Func :  ( 'brown' '(' )( expression )? ')';
lightGray_Func :  ( 'lightGray' '(' )( expression )? ')';
darkGray_Func :  ( 'darkGray' '(' )( expression )? ')';
lightBlue_Func :  ( 'lightBlue' '(' )( expression )? ')';
lightGreen_Func :  ( 'lightGreen' '(' )( expression )? ')';
lightCyan_Func :  ( 'lightCyan' '(' )( expression )? ')';
lightRed_Func :  ( 'lightRed' '(' )( expression )? ')';
lightMagenta_Func :  ( 'lightMagenta' '(' )( expression )? ')';
yellow_Func :  ( 'yellow' '(' )( expression )? ')';
white_Func :  ( 'white' '(' )( expression )? ')';
qlikTechBlue_Func :  ( 'qlikTechBlue' '(' )( expression )? ')';
qlikTechGray_Func :  ( 'qlikTechGray' '(' )( expression )? ')';
rangeNpv_Func :  ( 'rangeNpv' '(' ) expression ',' expression (',' expression)*  ')';
rangeIrr_Func :  ( 'rangeIrr' '(' ) expression ',' expression (',' expression)*  ')';
rangeXnpv_Func :  ( 'rangeXnpv' '(' ) expression ',' expression ',' expression (',' expression)*  ')';
rangeXirr_Func :  ( 'rangeXirr' '(' ) expression ',' expression (',' expression)*  ')';
blackAndSchole_Func :  ( 'blackAndSchole' '(' ) expression ',' expression ',' expression ',' expression ',' expression ',' expression  ')';
rangeCorrel_Func :  ( 'rangeCorrel' '(' ) expression ',' expression (',' expression)*  ')';
chiDist_Func :  ( 'chiDist' '(' ) expression ',' expression  ')';
chiDensity_Func :  ( 'chiDensity' '(' ) expression ',' expression  ')';
chiInv_Func :  ( 'chiInv' '(' ) expression ',' expression  ')';
normDist_Func :  ( 'normDist' '(' ) expression (',' expression (',' expression (',' expression )? )?)? ')';
normInv_Func :  ( 'normInv' '(' ) expression ',' expression ',' expression  ')';
tDist_Func :  ( 'tDist' '(' ) expression ',' expression ',' expression  ')';
tDensity_Func :  ( 'tDensity' '(' ) expression ',' expression  ')';
tInv_Func :  ( 'tInv' '(' ) expression ',' expression  ')';
fDist_Func :  ( 'fDist' '(' ) expression ',' expression ',' expression  ')';
fDensity_Func :  ( 'fDensity' '(' ) expression ',' expression ',' expression  ')';
fInv_Func :  ( 'fInv' '(' ) expression ',' expression ',' expression  ')';
betaDist_Func :  ( 'betaDist' '(' ) expression ',' expression ',' expression  ')';
betaDensity_Func :  ( 'betaDensity' '(' ) expression ',' expression ',' expression  ')';
betaInv_Func :  ( 'betaInv' '(' ) expression ',' expression ',' expression  ')';
gammaDist_Func :  ( 'gammaDist' '(' ) expression ',' expression ',' expression  ')';
gammaDensity_Func :  ( 'gammaDensity' '(' ) expression ',' expression ',' expression  ')';
gammaInv_Func :  ( 'gammaInv' '(' ) expression ',' expression ',' expression  ')';
poissonDist_Func :  ( 'poissonDist' '(' ) expression ',' expression  ')';
poissonInv_Func :  ( 'poissonInv' '(' ) expression ',' expression  ')';
poissonDensity_Func :  ( 'poissonDensity' '(' ) expression ',' expression  ')';
poissonFrequency_Func :  ( 'poissonFrequency' '(' ) expression ',' expression  ')';
binomDist_Func :  ( 'binomDist' '(' ) expression ',' expression ',' expression  ')';
binomInv_Func :  ( 'binomInv' '(' ) expression ',' expression ',' expression  ')';
binomFrequency_Func :  ( 'binomFrequency' '(' ) expression ',' expression ',' expression  ')';
geoBoundingBox_Func :  ( 'geoBoundingBox' '(' ) expression  ')';
geoReduceGeometry_Func :  ( 'geoReduceGeometry' '(' ) expression (',' expression )? ')';
geoProjectGeometry_Func :  ( 'geoProjectGeometry' '(' ) expression ',' expression  ')';
geoInvProjectGeometry_Func :  ( 'geoInvProjectGeometry' '(' ) expression ',' expression  ')';
geoProject_Func :  ( 'geoProject' '(' ) expression ',' expression  ')';
geoGetBoundingBox_Func :  ( 'geoGetBoundingBox' '(' ) expression  ')';
geoGetPolygonCenter_Func :  ( 'geoGetPolygonCenter' '(' ) expression  ')';
geoMakePoint_Func :  ( 'geoMakePoint' '(' ) expression ',' expression  ')';
geoAggrGeometry_Func :  ( 'geoAggrGeometry' '(' ) expression  ')';
geoCountVertex_Func :  ( 'geoCountVertex' '(' ) expression  ')';
tTest_t_Func :  ( 'tTest_t' '(' ) expression ',' expression (',' expression )? ')';
tTestw_t_Func :  ( 'tTestw_t' '(' ) expression ',' expression ',' expression (',' expression (',' expression )?)? ')';
tTest_DF_Func :  ( 'tTest_DF' '(' ) expression ',' expression (',' expression )? ')';
tTestw_DF_Func :  ( 'tTestw_DF' '(' ) expression ',' expression ',' expression (',' expression (',' expression )?)? ')';
tTest_Sig_Func :  ( 'tTest_Sig' '(' ) expression ',' expression (',' expression )? ')';
tTestw_Sig_Func :  ( 'tTestw_Sig' '(' ) expression ',' expression ',' expression (',' expression (',' expression )?)? ')';
tTest_Dif_Func :  ( 'tTest_Dif' '(' ) expression ',' expression  ')';
tTestw_Dif_Func :  ( 'tTestw_Dif' '(' ) expression ',' expression ',' expression  ')';
tTest_StErr_Func :  ( 'tTest_StErr' '(' ) expression ',' expression (',' expression )? ')';
tTestw_StErr_Func :  ( 'tTestw_StErr' '(' ) expression ',' expression ',' expression (',' expression (',' expression )?)? ')';
tTest_Conf_Func :  ( 'tTest_Conf' '(' ) expression ',' expression (',' expression (',' expression )?)? ')';
tTestw_Conf_Func :  ( 'tTestw_Conf' '(' ) expression ',' expression ',' expression (',' expression (',' expression (',' expression )? )?)? ')';
tTest_Lower_Func :  ( 'tTest_Lower' '(' ) expression ',' expression (',' expression (',' expression )?)? ')';
tTestw_Lower_Func :  ( 'tTestw_Lower' '(' ) expression ',' expression ',' expression (',' expression (',' expression (',' expression )? )?)? ')';
tTest_Upper_Func :  ( 'tTest_Upper' '(' ) expression ',' expression (',' expression (',' expression )?)? ')';
tTestw_Upper_Func :  ( 'tTestw_Upper' '(' ) expression ',' expression ',' expression (',' expression (',' expression (',' expression )? )?)? ')';
tTest1_t_Func :  ( 'tTest1_t' '(' ) expression  ')';
tTest1w_t_Func :  ( 'tTest1w_t' '(' ) expression ',' expression (',' expression )? ')';
tTest1_DF_Func :  ( 'tTest1_DF' '(' ) expression  ')';
tTest1w_DF_Func :  ( 'tTest1w_DF' '(' ) expression ',' expression (',' expression )? ')';
tTest1_Sig_Func :  ( 'tTest1_Sig' '(' ) expression  ')';
tTest1w_Sig_Func :  ( 'tTest1w_Sig' '(' ) expression ',' expression (',' expression )? ')';
tTest1_Dif_Func :  ( 'tTest1_Dif' '(' ) expression  ')';
tTest1w_Dif_Func :  ( 'tTest1w_Dif' '(' ) expression ',' expression  ')';
tTest1_StErr_Func :  ( 'tTest1_StErr' '(' ) expression  ')';
tTest1w_StErr_Func :  ( 'tTest1w_StErr' '(' ) expression ',' expression (',' expression )? ')';
tTest1_Conf_Func :  ( 'tTest1_Conf' '(' ) expression (',' expression )? ')';
tTest1w_Conf_Func :  ( 'tTest1w_Conf' '(' ) expression ',' expression (',' expression (',' expression )?)? ')';
tTest1_Lower_Func :  ( 'tTest1_Lower' '(' ) expression (',' expression )? ')';
tTest1w_Lower_Func :  ( 'tTest1w_Lower' '(' ) expression ',' expression (',' expression (',' expression )?)? ')';
tTest1_Upper_Func :  ( 'tTest1_Upper' '(' ) expression (',' expression )? ')';
tTest1w_Upper_Func :  ( 'tTest1w_Upper' '(' ) expression ',' expression (',' expression (',' expression )?)? ')';
zTest_z_Func :  ( 'zTest_z' '(' ) expression (',' expression )? ')';
zTestw_z_Func :  ( 'zTestw_z' '(' ) expression ',' expression (',' expression (',' expression )?)? ')';
zTest_Sig_Func :  ( 'zTest_Sig' '(' ) expression (',' expression )? ')';
zTestw_Sig_Func :  ( 'zTestw_Sig' '(' ) expression ',' expression (',' expression (',' expression )?)? ')';
zTest_Dif_Func :  ( 'zTest_Dif' '(' ) expression (',' expression )? ')';
zTestw_Dif_Func :  ( 'zTestw_Dif' '(' ) expression ',' expression (',' expression )? ')';
zTest_StErr_Func :  ( 'zTest_StErr' '(' ) expression (',' expression )? ')';
zTestw_StErr_Func :  ( 'zTestw_StErr' '(' ) expression ',' expression (',' expression (',' expression )?)? ')';
zTest_Conf_Func :  ( 'zTest_Conf' '(' ) expression (',' expression (',' expression )?)? ')';
zTestw_Conf_Func :  ( 'zTestw_Conf' '(' ) expression ',' expression (',' expression (',' expression (',' expression )? )?)? ')';
zTest_Lower_Func :  ( 'zTest_Lower' '(' ) expression (',' expression (',' expression )?)? ')';
zTestw_Lower_Func :  ( 'zTestw_Lower' '(' ) expression ',' expression (',' expression (',' expression (',' expression )? )?)? ')';
zTest_Upper_Func :  ( 'zTest_Upper' '(' ) expression (',' expression (',' expression )?)? ')';
zTestw_Upper_Func :  ( 'zTestw_Upper' '(' ) expression ',' expression (',' expression (',' expression (',' expression )? )?)? ')';
chi2Test_p_Func :  ( 'chi2Test_p' '(' ) expression ',' expression ',' expression (',' expression )? ')';
chi2Test_DF_Func :  ( 'chi2Test_DF' '(' ) expression ',' expression ',' expression (',' expression )? ')';
chi2Test_Chi2_Func :  ( 'chi2Test_Chi2' '(' ) expression ',' expression ',' expression (',' expression )? ')';
if_Func :  ( 'if' '(' ) expression ',' expression (',' expression )? ')';
pick_Func :  ( 'pick' '(' ) expression (',' expression)*  ')';
match_Func :  ( 'match' '(' ) expression ',' expression (',' expression)*  ')';
mixMatch_Func :  ( 'mixMatch' '(' ) expression ',' expression (',' expression)*  ')';
wildMatch_Func :  ( 'wildMatch' '(' ) expression ',' expression (',' expression)*  ')';
fastMatch_Func :  ( 'fastMatch' '(' ) expression ',' expression (',' expression)*  ')';
alt_Func :  ( 'alt' '(' ) expression (',' expression)*  ')';
coalesce_Func :  ( 'coalesce' '(' ) expression (',' expression)*  ')';
//emptyIsNull_Func :  ( 'emptyIsNull' '(' ) expression  ')';
trim_Func :  ( 'trim' '(' ) expression  ')';
lTrim_Func :  ( 'lTrim' '(' ) expression  ')';
rTrim_Func :  ( 'rTrim' '(' ) expression  ')';
purgeChar_Func :  ( 'purgeChar' '(' ) expression ',' expression  ')';
levenshteinDist_Func :  ( 'levenshteinDist' '(' ) expression ',' expression  ')';
keepChar_Func :  ( 'keepChar' '(' ) expression ',' expression  ')';
left_Func :  ( 'left' '(' ) expression ',' expression  ')';
right_Func :  ( 'right' '(' ) expression ',' expression  ')';
mid_Func :  ( 'mid' '(' ) expression ',' expression (',' expression )? ')';
subField_Func :  ( 'subField' '(' ) expression ',' expression (',' expression )? ')';
textBetween_Func :  ( 'textBetween' '(' ) expression ',' expression ',' expression (',' expression )? ')';
repeat_Func :  ( 'repeat' '(' ) expression ',' expression  ')';
chr_Func :  ( 'chr' '(' ) expression  ')';
class_Func :  ( 'class' '(' ) expression ',' expression (',' expression (',' expression )?)? ')';
index_Func :  ( 'index' '(' ) expression ',' expression (',' expression )? ')';
findOneOf_Func :  ( 'findOneOf' '(' ) expression ',' expression (',' expression )? ')';
len_Func :  ( 'len' '(' ) expression  ')';
isJson_Func :  ( 'isJson' '(' ) expression (',' expression )? ')';
jsonGet_Func :  ( 'jsonGet' '(' ) expression ',' expression  ')';
jsonSet_Func :  ( 'jsonSet' '(' ) expression ',' expression ',' expression  ')';
mod_Func :  ( 'mod' '(' ) expression ',' expression  ')';
odd_Func :  ( 'odd' '(' ) expression  ')';
even_Func :  ( 'even' '(' ) expression  ')';
uTC_Func :  ( 'uTC' '(' )( expression )? ')';
gMT_Func :  ( 'gMT' '(' )( expression )? ')';
now_Func :  ( 'now' '(' )( expression )? ')';
today_Func :  ( 'today' '(' )( expression )? ')';
elapsedSeconds_Func :  ( 'elapsedSeconds' '(' ) ')';
hour_Func :  ( 'hour' '(' ) expression  ')';
minute_Func :  ( 'minute' '(' ) expression  ')';
second_Func :  ( 'second' '(' ) expression  ')';
frac_Func :  ( 'frac' '(' ) expression  ')';
div_Func :  ( 'div' '(' ) expression ',' expression  ')';
distinct_on_Func: 'distinct on' '(' parameterList ')' ;
dual_Func :  ( 'dual' '(' ) expression ',' expression  ')';
bitCount_Func :  ( 'bitCount' '(' ) expression  ')';
ord_Func :  ( 'ord' '(' ) expression  ')';
upper_Func :  ( 'upper' '(' ) expression  ')';
applyCodepage_Func :  ( 'applyCodepage' '(' ) expression ',' expression  ')';
lower_Func :  ( 'lower' '(' ) expression  ')';
capitalize_Func :  ( 'capitalize' '(' ) expression  ')';
rGB_Func :  ( 'rGB' '(' ) expression ',' expression ',' expression  ')';
aRGB_Func :  ( 'aRGB' '(' ) expression ',' expression ',' expression ',' expression  ')';
hSL_Func :  ( 'hSL' '(' ) expression ',' expression ',' expression  ')';
color_Func :  ( 'color' '(' ) expression  ')';
sysColor_Func :  ( 'sysColor' '(' ) expression  ')';
colorMix1_Func :  ( 'colorMix1' '(' ) expression ',' expression ',' expression  ')';
colorMix2_Func :  ( 'colorMix2' '(' ) expression ',' expression ',' expression (',' expression )? ')';
colorMapJet_Func :  ( 'colorMapJet' '(' ) expression  ')';
colorMapHue_Func :  ( 'colorMapHue' '(' ) expression  ')';
isNull_Func :  ( 'isNull' '(' ) expression  ')';
isNum_Func :  ( 'isNum' '(' ) expression  ')';
isText_Func :  ( 'isText' '(' ) expression  ')';
recNo_Func :  ( 'recNo' '(' ) ')';
iterNo_Func :  ( 'iterNo' '(' ) ')';
noOfRows_Func :  ( 'noOfRows' '(' ) expression  ')';
noOfFields_Func :  ( 'noOfFields' '(' ) expression  ')';
fieldName_Func :  ( 'fieldName' '(' ) expression ',' expression  ')';
fieldNumber_Func :  ( 'fieldNumber' '(' ) expression ',' expression  ')';
rowNo_Func :  ( 'rowNo' '(' )('total')? ')';
timestamp_Func :  ( 'timestamp' '(' ) expression (',' expression )? ')';
date_Func :  ( 'date' '(' ) parameterList ')';
time_Func :  ( 'time' '(' ) expression (',' expression )? ')';
interval_Func :  ( 'interval' '(' ) expression (',' expression )? ')';
num_Func :  ( 'num' '(' ) expression (',' expression (',' expression (',' expression )? )?)? ')';
money_Func :  ( 'money' '(' ) expression (',' expression (',' expression (',' expression )? )?)? ')';
text_Func :  ( 'text' '(' ) expression  ')';
timestampH_Func :  ( 'Timestamp#' '(' ) expression (',' expression )? ')';
dateH_Func :  ( 'Date#' '(' ) parameterList ')';
timeH_Func :  ( 'Time#' '(' ) expression (',' expression )? ')';
intervalH_Func :  ( 'Interval#' '(' ) expression (',' expression )? ')';
numH_Func :  ( 'Num#' '(' ) expression (',' expression (',' expression (',' expression )? )?)? ')';
moneyH_Func :  ( 'Money#' '(' ) expression (',' expression (',' expression (',' expression )? )?)? ')';
month_Func :  ( 'month' '(' ) expression  ')';
day_Func :  ( 'day' '(' ) expression  ')';
week_Func :  ( 'week' '(' ) expression (',' expression (',' expression (',' expression )? )?)? ')';
weekDay_Func :  ( 'weekDay' '(' ) expression (',' expression )? ')';
weekYear_Func :  ( 'weekYear' '(' ) expression (',' expression (',' expression (',' expression )? )?)? ')';
year_Func :  ( 'year' '(' ) expression  ')';
age_Func :  ( 'age' '(' ) expression ',' expression  ')';
netWorkDays_Func :  ( 'netWorkDays' '(' ) expression ',' expression (',' expression (',' expression)* )? ')';
lastWorkDate_Func :  ( 'lastWorkDate' '(' ) expression ',' expression (',' expression (',' expression)* )? ')';
firstWorkDate_Func :  ( 'firstWorkDate' '(' ) expression ',' expression (',' expression (',' expression)* )? ')';
makeDate_Func :  ( 'makeDate' '(' ) expression (',' expression (',' expression )?)? ')';
makeWeekDate_Func :  ( 'makeWeekDate' '(' ) expression (',' expression (',' expression (',' expression (',' expression (',' expression )? )? )? )?)? ')';
addMonths_Func :  ( 'addMonths' '(' ) expression ',' expression (',' expression )? ')';
addYears_Func :  ( 'addYears' '(' ) expression ',' expression  ')';
makeTime_Func :  ( 'makeTime' '(' ) expression (',' expression (',' expression )?)? ')';
numSum_Func :  ( 'numSum' '(' ) expression (',' expression)*  ')';
numMin_Func :  ( 'numMin' '(' ) expression (',' expression)*  ')';
numMax_Func :  ( 'numMax' '(' ) expression (',' expression)*  ')';
numAvg_Func :  ( 'numAvg' '(' ) expression (',' expression)*  ')';
numCount_Func :  ( 'numCount' '(' ) expression (',' expression)*  ')';
rangeNumericCount_Func :  ( 'rangeNumericCount' '(' ) expression (',' expression)*  ')';
rangeSum_Func :  ( 'rangeSum' '(' ) expression (',' expression)*  ')';
rangeMin_Func :  ( 'rangeMin' '(' ) expression (',' expression)*  ')';
rangeMax_Func :  ( 'rangeMax' '(' ) expression (',' expression)*  ')';
rangeAvg_Func :  ( 'rangeAvg' '(' ) expression (',' expression)*  ')';
rangeStDev_Func :  ( 'rangeStDev' '(' ) expression (',' expression)*  ')';
rangeSkew_Func :  ( 'rangeSkew' '(' ) expression (',' expression)*  ')';
rangeKurtosis_Func :  ( 'rangeKurtosis' '(' ) expression (',' expression)*  ')';
rangeNullCount_Func :  ( 'rangeNullCount' '(' ) expression (',' expression)*  ')';
rangeTextCount_Func :  ( 'rangeTextCount' '(' ) expression (',' expression)*  ')';
rangeMissingCount_Func :  ( 'rangeMissingCount' '(' ) expression (',' expression)*  ')';
rangeCount_Func :  ( 'rangeCount' '(' ) expression (',' expression)*  ')';
rangeOnly_Func :  ( 'rangeOnly' '(' ) expression (',' expression)*  ')';
rangeMinString_Func :  ( 'rangeMinString' '(' ) expression (',' expression)*  ')';
rangeMaxString_Func :  ( 'rangeMaxString' '(' ) expression (',' expression)*  ')';
rangeMode_Func :  ( 'rangeMode' '(' ) expression (',' expression)*  ')';
rangeFractile_Func :  ( 'rangeFractile' '(' ) expression ',' expression (',' expression)*  ')';
rangeFractileExc_Func :  ( 'rangeFractileExc' '(' ) expression ',' expression (',' expression)*  ')';
previous_Func :  ( 'previous' '(' ) expression  ')';
peek_Func :  ( 'peek' '(' ) expression (',' expression (',' expression )?)? ')';
lookup_Func :  ( 'lookup' '(' ) expression ',' expression ',' expression (',' expression )? ')';
exists_Func :  ( 'exists' '(' ) expression (',' expression )? ')';
fieldValue_Func :  ( 'fieldValue' '(' ) expression ',' expression  ')';
fieldValueCount_Func :  ( 'fieldValueCount' '(' ) expression  ')';
fieldIndex_Func :  ( 'fieldIndex' '(' ) expression ',' expression  ')';
fieldElemNo_Func :  ( 'fieldElemNo' '(' ) expression  ')';
autoNumber_Func :  ( 'autoNumber' '(' ) expression (',' expression )? ')';
autoNumberHash128_Func :  ( 'autoNumberHash128' '(' ) expression (',' expression)*  ')';
autoNumberHash256_Func :  ( 'autoNumberHash256' '(' ) expression (',' expression)*  ')';
hash128_Func :  ( 'hash128' '(' ) expression (',' expression)*  ')';
hash160_Func :  ( 'hash160' '(' ) expression (',' expression)*  ')';
hash256_Func :  ( 'hash256' '(' ) expression (',' expression)*  ')';
applyMap_Func :  ( 'applyMap' '(' ) expression ',' expression (',' expression )? ')';
mapSubString_Func :  ( 'mapSubString' '(' ) expression ',' expression  ')';
replace_Func :  ( 'replace' '(' ) expression ',' expression ',' expression  ')';
subStringCount_Func :  ( 'subStringCount' '(' ) expression ',' expression  ')';
evaluate_Func :  ( 'evaluate' '(' ) expression  ')';
oSUser_Func :  ( 'oSUser' '(' ) ')';
getDataModelHash_Func :  ( 'getDataModelHash' '(' ) ')';
documentPath_Func :  ( 'documentPath' '(' ) ')';
documentName_Func :  ( 'documentName' '(' ) ')';
documentTitle_Func :  ( 'documentTitle' '(' ) ')';
noOfTables_Func :  ( 'noOfTables' '(' ) ')';
tableName_Func :  ( 'tableName' '(' ) expression  ')';
tableNumber_Func :  ( 'tableNumber' '(' ) expression  ')';
getCollationLocale_Func :  ( 'getCollationLocale' '(' ) ')';
qlikViewVersion_Func :  ( 'qlikViewVersion' '(' ) ')';
productVersion_Func :  ( 'productVersion' '(' ) ')';
engineVersion_Func :  ( 'engineVersion' '(' ) ')';
qVUser_Func :  ( 'qVUser' '(' ) ')';
computerName_Func :  ( 'computerName' '(' ) ')';
author_Func :  ( 'author' '(' ) ')';
reloadTime_Func :  ( 'reloadTime' '(' ) ')';
getObjectField_Func :  ( 'getObjectField' '(' )( expression (',' expression )?)? ')';
clientPlatform_Func :  ( 'clientPlatform' '(' ) ')';
dayNumberOfYear_Func :  ( 'dayNumberOfYear' '(' ) expression (',' expression )? ')';
dayNumberOfQuarter_Func :  ( 'dayNumberOfQuarter' '(' ) expression (',' expression )? ')';
year2Date_Func :  ( 'year2Date' '(' ) expression (',' expression (',' expression (',' expression )? )?)? ')';
yearToDate_Func :  ( 'yearToDate' '(' ) expression (',' expression (',' expression (',' expression )? )?)? ')';
inYear_Func :  ( 'inYear' '(' ) expression ',' expression ',' expression (',' expression )? ')';
inYearToDate_Func :  ( 'inYearToDate' '(' ) expression ',' expression ',' expression (',' expression )? ')';
inQuarter_Func :  ( 'inQuarter' '(' ) expression ',' expression ',' expression (',' expression )? ')';
inQuarterToDate_Func :  ( 'inQuarterToDate' '(' ) expression ',' expression ',' expression (',' expression )? ')';
inMonth_Func :  ( 'inMonth' '(' ) expression ',' expression ',' expression (',' expression )? ')';
inMonthToDate_Func :  ( 'inMonthToDate' '(' ) expression ',' expression ',' expression (',' expression )? ')';
inMonths_Func :  ( 'inMonths' '(' ) expression ',' expression ',' expression ',' expression (',' expression )? ')';
inMonthsToDate_Func :  ( 'inMonthsToDate' '(' ) expression ',' expression ',' expression ',' expression (',' expression )? ')';
inWeek_Func :  ( 'inWeek' '(' ) expression ',' expression ',' expression (',' expression )? ')';
inWeekToDate_Func :  ( 'inWeekToDate' '(' ) expression ',' expression ',' expression (',' expression )? ')';
inLunarWeek_Func :  ( 'inLunarWeek' '(' ) expression ',' expression ',' expression (',' expression )? ')';
inLunarWeekToDate_Func :  ( 'inLunarWeekToDate' '(' ) expression ',' expression ',' expression (',' expression )? ')';
inDay_Func :  ( 'inDay' '(' ) expression ',' expression ',' expression (',' expression )? ')';
inDayToTime_Func :  ( 'inDayToTime' '(' ) expression ',' expression ',' expression (',' expression )? ')';
yearStart_Func :  ( 'yearStart' '(' ) expression (',' expression (',' expression )?)? ')';
quarterStart_Func :  ( 'quarterStart' '(' ) expression (',' expression (',' expression )?)? ')';
monthStart_Func :  ( 'monthStart' '(' ) expression (',' expression )? ')';
monthsStart_Func :  ( 'monthsStart' '(' ) expression ',' expression (',' expression (',' expression )?)? ')';
weekStart_Func :  ( 'weekStart' '(' ) expression (',' expression (',' expression )?)? ')';
lunarWeekStart_Func :  ( 'lunarWeekStart' '(' ) expression (',' expression (',' expression )?)? ')';
dayStart_Func :  ( 'dayStart' '(' ) expression (',' expression (',' expression )?)? ')';
yearEnd_Func :  ( 'yearEnd' '(' ) expression (',' expression (',' expression )?)? ')';
quarterEnd_Func :  ( 'quarterEnd' '(' ) expression (',' expression (',' expression )?)? ')';
monthEnd_Func :  ( 'monthEnd' '(' ) expression (',' expression )? ')';
monthsEnd_Func :  ( 'monthsEnd' '(' ) expression ',' expression (',' expression (',' expression )?)? ')';
weekEnd_Func :  ( 'weekEnd' '(' ) expression (',' expression (',' expression )?)? ')';
lunarWeekEnd_Func :  ( 'lunarWeekEnd' '(' ) expression (',' expression (',' expression )?)? ')';
dayEnd_Func :  ( 'dayEnd' '(' ) expression (',' expression (',' expression )?)? ')';
yearName_Func :  ( 'yearName' '(' ) expression (',' expression (',' expression )?)? ')';
quarterName_Func :  ( 'quarterName' '(' ) expression (',' expression (',' expression )?)? ')';
monthName_Func :  ( 'monthName' '(' ) expression (',' expression )? ')';
monthsName_Func :  ( 'monthsName' '(' ) expression ',' expression (',' expression (',' expression )?)? ')';
weekName_Func :  ( 'weekName' '(' ) expression (',' expression (',' expression (',' expression (',' expression )? )? )?)? ')';
lunarWeekName_Func :  ( 'lunarWeekName' '(' ) expression (',' expression (',' expression )?)? ')';
dayName_Func :  ( 'dayName' '(' ) expression (',' expression (',' expression )?)? ')';
setDateYear_Func :  ( 'setDateYear' '(' ) expression ',' expression  ')';
setDateYearMonth_Func :  ( 'setDateYearMonth' '(' ) expression ',' expression ',' expression  ')';
localTime_Func :  ( 'localTime' '(' )( expression (',' expression )?)? ')';
convertToLocalTime_Func :  ( 'convertToLocalTime' '(' ) expression (',' expression (',' expression )?)? ')';
qvdCreateTime_Func :  ( 'qvdCreateTime' '(' ) expression  ')';
qvdNoOfRecords_Func :  ( 'qvdNoOfRecords' '(' ) expression  ')';
qvdNoOfFields_Func :  ( 'qvdNoOfFields' '(' ) expression  ')';
qvdFieldName_Func :  ( 'qvdFieldName' '(' ) expression ',' expression  ')';
qvdTableName_Func :  ( 'qvdTableName' '(' ) expression  ')';
fileTime_Func :  ( 'fileTime' '(' )( expression )? ')';
fileSize_Func :  ( 'fileSize' '(' )( expression )? ')';
attribute_Func :  ( 'attribute' '(' ) expression ',' expression  ')';
filePath_Func :  ( 'filePath' '(' )( expression )? ')';
fileName_Func :  ( 'fileName' '(' )( expression )? ')';
fileDir_Func :  ( 'fileDir' '(' )( expression )? ')';
getFolderPath_Func :  ( 'getFolderPath' '(' )( expression )? ')';
fileBaseName_Func :  ( 'fileBaseName' '(' )( expression )? ')';
fileExtension_Func :  ( 'fileExtension' '(' )( expression )? ')';
isPartialReload_Func :  ( 'isPartialReload' '(' ) ')';
connectString_Func :  ( 'connectString' '(' ) ')';
sum_Func :  ( 'sum' '(' )('distinct')? expression  ')';
min_Func :  ( 'min' '(' ) expression (',' expression )? ')';
max_Func :  ( 'max' '(' ) expression (',' expression )? ')';
avg_Func :  ( 'avg' '(' )('distinct')? expression  ')';
stDev_Func :  ( 'stDev' '(' )('distinct')? expression  ')';
skew_Func :  ( 'skew' '(' )('distinct')? expression  ')';
kurtosis_Func :  ( 'kurtosis' '(' )('distinct')? expression  ')';
numericCount_Func :  ( 'numericCount' '(' )('distinct')? expression  ')';
nullCount_Func :  ( 'nullCount' '(' )('distinct')? expression  ')';
textCount_Func :  ( 'textCount' '(' )('distinct')? expression  ')';
count_Func :  ( 'count' '(' )('distinct')? expression  ')';
missingCount_Func :  ( 'missingCount' '(' )('distinct')? expression  ')';
minString_Func :  ( 'minString' '(' ) expression  ')';
maxString_Func :  ( 'maxString' '(' ) expression  ')';
only_Func :  ( 'only' '(' ) expression  ')';
mode_Func :  ( 'mode' '(' ) expression  ')';
fractile_Func :  ( 'fractile' '(' ) ('total')?('all')? ('distinct')? expression ',' expression  ')';
median_Func :  ( 'median' '(' ) ('total')?('all')? ('distinct')? expression  ')';
firstValue_Func :  ( 'firstValue' '(' ) expression  ')';
lastValue_Func :  ( 'lastValue' '(' ) expression  ')';
stErr_Func :  ( 'stErr' '(' )('distinct')? expression  ')';
fractileExc_Func :  ( 'fractileExc' '(' ) ('total')?('all')? ('distinct')? expression ',' expression  ')';
concat_Func :  ( 'concat' '(' )('distinct')? ('total')? expression  (',' expression  (',' expression )?)? ')';
firstSortedValue_Func :  ( 'firstSortedValue' '(' ) ('distinct')? ('all')? ('total')? expression ',' expression (',' expression )? ')';
npv_Func :  ( 'npv' '(' ) expression ',' expression  ')';
irr_Func :  ( 'irr' '(' ) expression  ')';
xnpv_Func :  ( 'xnpv' '(' ) expression ',' expression ',' expression  ')';
xirr_Func :  ( 'xirr' '(' ) expression ',' expression  ')';
correl_Func :  ( 'correl' '(' )('total')? expression ',' expression  ')';
stEYX_Func :  ( 'stEYX' '(' )('total')? expression ',' expression  ')';
linEst_B_Func :  ( 'linEst_B' '(' ) ('total')? expression ',' expression (',' expression  (',' expression )?)? ')';
linEst_DF_Func :  ( 'linEst_DF' '(' ) ('total')? expression ',' expression (',' expression  (',' expression )?)? ')';
linEst_F_Func :  ( 'linEst_F' '(' )('total')? expression ',' expression (',' expression  (',' expression )?)? ')';
linEst_M_Func :  ( 'linEst_M' '(' ) ('total')? expression ',' expression (',' expression  (',' expression )?)? ')';
linEst_R2_Func :  ( 'linEst_R2' '(' ) ('total')? expression ',' expression (',' expression  (',' expression )?)? ')';
linEst_SEB_Func :  ( 'linEst_SEB' '(' ) ('total')? expression ',' expression (',' expression  (',' expression )?)? ')';
linEst_SEM_Func :  ( 'linEst_SEM' '(' ) ('total')? expression ',' expression (',' expression  (',' expression )?)? ')';
linEst_SEY_Func :  ( 'linEst_SEY' '(' ) ('total')? expression ',' expression (',' expression  (',' expression )?)? ')';
linEst_SSReg_Func :  ( 'linEst_SSReg' '(' ) ('total')? expression ',' expression (',' expression (',' expression )? )? ')';
linEst_SSResid_Func :  ( 'linEst_SSResid' '(' ) ('total')? expression ',' expression (',' expression  (',' expression )?)? ')';
hCValue_Func :  ( 'hCValue' '(' ) expression ',' expression  ')';
hCNoRows_Func :  ( 'hCNoRows' '(' ) ')';
case_Func: 'case' ('when' expression 'then' expression)* ('else' expression)? 'end';
otherUnknownFuncCall: Identifier '(' ('distinct')? ('total')? ('all')? parameterList ')';