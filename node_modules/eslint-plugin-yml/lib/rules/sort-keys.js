"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const natural_compare_1 = __importDefault(require("natural-compare"));
const index_1 = require("../utils/index");
const ast_utils_1 = require("../utils/ast-utils");
const compat_1 = require("../utils/compat");
function isNewLine(char) {
    return (char === "\n" || char === "\r" || char === "\u2028" || char === "\u2029");
}
function getPropertyName(node, sourceCode) {
    const prop = node.key;
    if (prop == null) {
        return "";
    }
    const target = prop.type === "YAMLWithMeta" ? prop.value : prop;
    if (target == null) {
        return "";
    }
    if (target.type === "YAMLScalar" && typeof target.value === "string") {
        return target.value;
    }
    return sourceCode.text.slice(...target.range);
}
class YAMLPairData {
    get reportLoc() {
        var _a, _b;
        return (_b = (_a = this.node.key) === null || _a === void 0 ? void 0 : _a.loc) !== null && _b !== void 0 ? _b : this.node.loc;
    }
    constructor(mapping, node, index, anchorAlias) {
        this.cachedName = null;
        this.mapping = mapping;
        this.node = node;
        this.index = index;
        this.anchorAlias = anchorAlias;
    }
    get name() {
        var _a;
        return ((_a = this.cachedName) !== null && _a !== void 0 ? _a : (this.cachedName = getPropertyName(this.node, this.mapping.sourceCode)));
    }
    getPrev() {
        const prevIndex = this.index - 1;
        return prevIndex >= 0 ? this.mapping.pairs[prevIndex] : null;
    }
}
class YAMLMappingData {
    constructor(node, sourceCode, anchorAliasMap) {
        this.cachedProperties = null;
        this.node = node;
        this.sourceCode = sourceCode;
        this.anchorAliasMap = anchorAliasMap;
    }
    get pairs() {
        var _a;
        return ((_a = this.cachedProperties) !== null && _a !== void 0 ? _a : (this.cachedProperties = this.node.pairs.map((e, index) => new YAMLPairData(this, e, index, this.anchorAliasMap.get(e)))));
    }
    getPath(sourceCode) {
        let path = "";
        let curr = this.node;
        let p = curr.parent;
        while (p) {
            if (p.type === "YAMLPair") {
                const name = getPropertyName(p, sourceCode);
                if (/^[$a-z_][\w$]*$/iu.test(name)) {
                    path = `.${name}${path}`;
                }
                else {
                    path = `[${JSON.stringify(name)}]${path}`;
                }
            }
            else if (p.type === "YAMLSequence") {
                const index = p.entries.indexOf(curr);
                path = `[${index}]${path}`;
            }
            curr = p;
            p = curr.parent;
        }
        if (path.startsWith(".")) {
            path = path.slice(1);
        }
        return path;
    }
}
function isCompatibleWithESLintOptions(options) {
    if (options.length === 0) {
        return true;
    }
    if (typeof options[0] === "string" || options[0] == null) {
        return true;
    }
    return false;
}
function buildValidatorFromType(order, insensitive, natural) {
    let compare = natural
        ? ([a, b]) => (0, natural_compare_1.default)(a, b) <= 0
        : ([a, b]) => a <= b;
    if (insensitive) {
        const baseCompare = compare;
        compare = ([a, b]) => baseCompare([a.toLowerCase(), b.toLowerCase()]);
    }
    if (order === "desc") {
        const baseCompare = compare;
        compare = (args) => baseCompare(args.reverse());
    }
    return (a, b) => compare([a.name, b.name]);
}
function parseOptions(options, sourceCode) {
    var _a, _b, _c;
    if (isCompatibleWithESLintOptions(options)) {
        const type = (_a = options[0]) !== null && _a !== void 0 ? _a : "asc";
        const obj = (_b = options[1]) !== null && _b !== void 0 ? _b : {};
        const insensitive = obj.caseSensitive === false;
        const natural = Boolean(obj.natural);
        const minKeys = (_c = obj.minKeys) !== null && _c !== void 0 ? _c : 2;
        const allowLineSeparatedGroups = obj.allowLineSeparatedGroups || false;
        return [
            {
                isTargetMapping: (data) => data.node.pairs.length >= minKeys,
                ignore: () => false,
                isValidOrder: buildValidatorFromType(type, insensitive, natural),
                orderText: `${natural ? "natural " : ""}${insensitive ? "insensitive " : ""}${type}ending`,
                allowLineSeparatedGroups,
            },
        ];
    }
    return options.map((opt) => {
        var _a, _b, _c, _d, _e;
        const order = opt.order;
        const pathPattern = new RegExp(opt.pathPattern);
        const hasProperties = (_a = opt.hasProperties) !== null && _a !== void 0 ? _a : [];
        const minKeys = (_b = opt.minKeys) !== null && _b !== void 0 ? _b : 2;
        const allowLineSeparatedGroups = opt.allowLineSeparatedGroups || false;
        if (!Array.isArray(order)) {
            const type = (_c = order.type) !== null && _c !== void 0 ? _c : "asc";
            const insensitive = order.caseSensitive === false;
            const natural = Boolean(order.natural);
            return {
                isTargetMapping,
                ignore: () => false,
                isValidOrder: buildValidatorFromType(type, insensitive, natural),
                orderText: `${natural ? "natural " : ""}${insensitive ? "insensitive " : ""}${type}ending`,
                allowLineSeparatedGroups,
            };
        }
        const parsedOrder = [];
        for (const o of order) {
            if (typeof o === "string") {
                parsedOrder.push({
                    test: (data) => data.name === o,
                    isValidNestOrder: () => true,
                });
            }
            else {
                const keyPattern = o.keyPattern ? new RegExp(o.keyPattern) : null;
                const nestOrder = (_d = o.order) !== null && _d !== void 0 ? _d : {};
                const type = (_e = nestOrder.type) !== null && _e !== void 0 ? _e : "asc";
                const insensitive = nestOrder.caseSensitive === false;
                const natural = Boolean(nestOrder.natural);
                parsedOrder.push({
                    test: (data) => (keyPattern ? keyPattern.test(data.name) : true),
                    isValidNestOrder: buildValidatorFromType(type, insensitive, natural),
                });
            }
        }
        return {
            isTargetMapping,
            ignore: (data) => parsedOrder.every((p) => !p.test(data)),
            isValidOrder(a, b) {
                for (const p of parsedOrder) {
                    const matchA = p.test(a);
                    const matchB = p.test(b);
                    if (!matchA || !matchB) {
                        if (matchA) {
                            return true;
                        }
                        if (matchB) {
                            return false;
                        }
                        continue;
                    }
                    return p.isValidNestOrder(a, b);
                }
                return false;
            },
            orderText: "specified",
            allowLineSeparatedGroups,
        };
        function isTargetMapping(data) {
            if (data.node.pairs.length < minKeys) {
                return false;
            }
            if (hasProperties.length > 0) {
                const names = new Set(data.pairs.map((p) => p.name));
                if (!hasProperties.every((name) => names.has(name))) {
                    return false;
                }
            }
            return pathPattern.test(data.getPath(sourceCode));
        }
    });
}
const ALLOW_ORDER_TYPES = ["asc", "desc"];
const ORDER_OBJECT_SCHEMA = {
    type: "object",
    properties: {
        type: {
            enum: ALLOW_ORDER_TYPES,
        },
        caseSensitive: {
            type: "boolean",
        },
        natural: {
            type: "boolean",
        },
    },
    additionalProperties: false,
};
exports.default = (0, index_1.createRule)("sort-keys", {
    meta: {
        docs: {
            description: "require mapping keys to be sorted",
            categories: null,
            extensionRule: false,
            layout: false,
        },
        fixable: "code",
        schema: {
            oneOf: [
                {
                    type: "array",
                    items: {
                        type: "object",
                        properties: {
                            pathPattern: { type: "string" },
                            hasProperties: {
                                type: "array",
                                items: { type: "string" },
                            },
                            order: {
                                oneOf: [
                                    {
                                        type: "array",
                                        items: {
                                            anyOf: [
                                                { type: "string" },
                                                {
                                                    type: "object",
                                                    properties: {
                                                        keyPattern: {
                                                            type: "string",
                                                        },
                                                        order: ORDER_OBJECT_SCHEMA,
                                                    },
                                                    additionalProperties: false,
                                                },
                                            ],
                                        },
                                        uniqueItems: true,
                                    },
                                    ORDER_OBJECT_SCHEMA,
                                ],
                            },
                            minKeys: {
                                type: "integer",
                                minimum: 2,
                            },
                            allowLineSeparatedGroups: {
                                type: "boolean",
                            },
                        },
                        required: ["pathPattern", "order"],
                        additionalProperties: false,
                    },
                    minItems: 1,
                },
                {
                    type: "array",
                    items: [
                        {
                            enum: ALLOW_ORDER_TYPES,
                        },
                        {
                            type: "object",
                            properties: {
                                caseSensitive: {
                                    type: "boolean",
                                },
                                natural: {
                                    type: "boolean",
                                },
                                minKeys: {
                                    type: "integer",
                                    minimum: 2,
                                },
                                allowLineSeparatedGroups: {
                                    type: "boolean",
                                },
                            },
                            additionalProperties: false,
                        },
                    ],
                    additionalItems: false,
                },
            ],
        },
        messages: {
            sortKeys: "Expected mapping keys to be in {{orderText}} order. '{{thisName}}' should be before '{{prevName}}'.",
        },
        type: "suggestion",
    },
    create(context) {
        var _a;
        const sourceCode = (0, compat_1.getSourceCode)(context);
        if (!((_a = sourceCode.parserServices) === null || _a === void 0 ? void 0 : _a.isYAML)) {
            return {};
        }
        const parsedOptions = parseOptions(context.options, sourceCode);
        function isValidOrder(prevData, thisData, option) {
            if (option.isValidOrder(prevData, thisData)) {
                return true;
            }
            for (const aliasName of thisData.anchorAlias.aliases) {
                if (prevData.anchorAlias.anchors.has(aliasName)) {
                    return true;
                }
            }
            for (const anchorName of thisData.anchorAlias.anchors) {
                if (prevData.anchorAlias.aliases.has(anchorName)) {
                    return true;
                }
            }
            return false;
        }
        function ignore(data, option) {
            if (!data.node.key && !data.node.value) {
                return true;
            }
            return option.ignore(data);
        }
        function verifyPair(data, option) {
            if (ignore(data, option)) {
                return;
            }
            const prevList = [];
            let currTarget = data;
            let prevTarget;
            while ((prevTarget = currTarget.getPrev())) {
                if (option.allowLineSeparatedGroups) {
                    if (hasBlankLine(prevTarget, currTarget)) {
                        break;
                    }
                }
                if (!ignore(prevTarget, option)) {
                    prevList.push(prevTarget);
                }
                currTarget = prevTarget;
            }
            if (prevList.length === 0) {
                return;
            }
            const prev = prevList[0];
            if (!isValidOrder(prev, data, option)) {
                context.report({
                    loc: data.reportLoc,
                    messageId: "sortKeys",
                    data: {
                        thisName: data.name,
                        prevName: prev.name,
                        orderText: option.orderText,
                    },
                    *fix(fixer) {
                        let moveTarget = prevList[0];
                        for (const prev of prevList) {
                            if (isValidOrder(prev, data, option)) {
                                break;
                            }
                            else {
                                moveTarget = prev;
                            }
                        }
                        if (data.mapping.node.style === "flow") {
                            yield* fixForFlow(fixer, data, moveTarget);
                        }
                        else {
                            yield* fixForBlock(fixer, data, moveTarget);
                        }
                    },
                });
            }
        }
        function hasBlankLine(prev, next) {
            const tokenOrNodes = [
                ...sourceCode.getTokensBetween(prev.node, next.node, {
                    includeComments: true,
                }),
                next.node,
            ];
            let prevLoc = prev.node.loc;
            for (const t of tokenOrNodes) {
                const loc = t.loc;
                if (loc.start.line - prevLoc.end.line > 1) {
                    return true;
                }
                prevLoc = loc;
            }
            return false;
        }
        let pairStack = {
            upper: null,
            anchors: new Set(),
            aliases: new Set(),
        };
        const anchorAliasMap = new Map();
        return {
            YAMLPair() {
                pairStack = {
                    upper: pairStack,
                    anchors: new Set(),
                    aliases: new Set(),
                };
            },
            YAMLAnchor(node) {
                if (pairStack) {
                    pairStack.anchors.add(node.name);
                }
            },
            YAMLAlias(node) {
                if (pairStack) {
                    pairStack.aliases.add(node.name);
                }
            },
            "YAMLPair:exit"(node) {
                anchorAliasMap.set(node, pairStack);
                const { anchors, aliases } = pairStack;
                pairStack = pairStack.upper;
                pairStack.anchors = new Set([...pairStack.anchors, ...anchors]);
                pairStack.aliases = new Set([...pairStack.aliases, ...aliases]);
            },
            "YAMLMapping:exit"(node) {
                const data = new YAMLMappingData(node, sourceCode, anchorAliasMap);
                const option = parsedOptions.find((o) => o.isTargetMapping(data));
                if (!option) {
                    return;
                }
                for (const pair of data.pairs) {
                    verifyPair(pair, option);
                }
            },
        };
        function* fixForFlow(fixer, data, moveTarget) {
            const beforeCommaToken = sourceCode.getTokenBefore(data.node);
            let insertCode, removeRange, insertTargetToken;
            const afterCommaToken = sourceCode.getTokenAfter(data.node);
            const moveTargetBeforeToken = sourceCode.getTokenBefore(moveTarget.node);
            if ((0, ast_utils_1.isComma)(afterCommaToken)) {
                removeRange = [beforeCommaToken.range[1], afterCommaToken.range[1]];
                insertCode = sourceCode.text.slice(...removeRange);
                insertTargetToken = moveTargetBeforeToken;
            }
            else {
                removeRange = [beforeCommaToken.range[0], data.node.range[1]];
                if ((0, ast_utils_1.isComma)(moveTargetBeforeToken)) {
                    insertCode = sourceCode.text.slice(...removeRange);
                    insertTargetToken = sourceCode.getTokenBefore(moveTargetBeforeToken);
                }
                else {
                    insertCode = `${sourceCode.text.slice(beforeCommaToken.range[1], data.node.range[1])},`;
                    insertTargetToken = moveTargetBeforeToken;
                }
            }
            yield fixer.insertTextAfterRange(insertTargetToken.range, insertCode);
            yield fixer.removeRange(removeRange);
        }
        function* fixForBlock(fixer, data, moveTarget) {
            const nodeLocs = getPairRangeForBlock(data.node);
            const moveTargetLocs = getPairRangeForBlock(moveTarget.node);
            if (moveTargetLocs.loc.start.column === 0) {
                const removeRange = [
                    getNewlineStartIndex(nodeLocs.range[0]),
                    nodeLocs.range[1],
                ];
                const moveTargetRange = [
                    getNewlineStartIndex(moveTargetLocs.range[0]),
                    moveTargetLocs.range[1],
                ];
                const insertCode = sourceCode.text.slice(...removeRange);
                yield fixer.insertTextBeforeRange(moveTargetRange, `${insertCode}${moveTargetLocs.loc.start.line === 1 ? "\n" : ""}`);
                yield fixer.removeRange(removeRange);
            }
            else {
                const diffIndent = nodeLocs.indentColumn - moveTargetLocs.indentColumn;
                const insertCode = `${sourceCode.text.slice(nodeLocs.range[0] + diffIndent, nodeLocs.range[1])}\n${sourceCode.text.slice(nodeLocs.range[0], nodeLocs.range[0] + diffIndent)}`;
                yield fixer.insertTextBeforeRange(moveTargetLocs.range, insertCode);
                const removeRange = [
                    getNewlineStartIndex(nodeLocs.range[0]),
                    nodeLocs.range[1],
                ];
                yield fixer.removeRange(removeRange);
            }
        }
        function getNewlineStartIndex(nextIndex) {
            for (let index = nextIndex; index >= 0; index--) {
                const char = sourceCode.text[index];
                if (isNewLine(sourceCode.text[index])) {
                    const prev = sourceCode.text[index - 1];
                    if (prev === "\r" && char === "\n") {
                        return index - 1;
                    }
                    return index;
                }
            }
            return 0;
        }
        function getPairRangeForBlock(node) {
            let endOfRange, end;
            const afterToken = sourceCode.getTokenAfter(node, {
                includeComments: true,
                filter: (t) => !(0, ast_utils_1.isCommentToken)(t) || node.loc.end.line < t.loc.start.line,
            });
            if (!afterToken || node.loc.end.line < afterToken.loc.start.line) {
                const line = afterToken
                    ? afterToken.loc.start.line - 1
                    : node.loc.end.line;
                const lineText = sourceCode.lines[line - 1];
                end = {
                    line,
                    column: lineText.length,
                };
                endOfRange = sourceCode.getIndexFromLoc(end);
            }
            else {
                endOfRange = node.range[1];
                end = node.loc.end;
            }
            const beforeToken = sourceCode.getTokenBefore(node);
            if (beforeToken) {
                const next = sourceCode.getTokenAfter(beforeToken, {
                    includeComments: true,
                });
                if (beforeToken.loc.end.line < next.loc.start.line ||
                    beforeToken.loc.end.line < node.loc.start.line) {
                    const start = {
                        line: beforeToken.loc.end.line < next.loc.start.line
                            ? next.loc.start.line
                            : node.loc.start.line,
                        column: 0,
                    };
                    const startOfRange = sourceCode.getIndexFromLoc(start);
                    return {
                        range: [startOfRange, endOfRange],
                        loc: { start, end },
                        indentColumn: next.loc.start.column,
                    };
                }
                const start = beforeToken.loc.end;
                const startOfRange = beforeToken.range[1];
                return {
                    range: [startOfRange, endOfRange],
                    loc: { start, end },
                    indentColumn: node.range[0] - beforeToken.range[1],
                };
            }
            let next = node;
            for (const beforeComment of sourceCode
                .getTokensBefore(node, {
                includeComments: true,
            })
                .reverse()) {
                if (beforeComment.loc.end.line + 1 < next.loc.start.line) {
                    const start = {
                        line: next.loc.start.line,
                        column: 0,
                    };
                    const startOfRange = sourceCode.getIndexFromLoc(start);
                    return {
                        range: [startOfRange, endOfRange],
                        loc: { start, end },
                        indentColumn: next.loc.start.column,
                    };
                }
                next = beforeComment;
            }
            const start = {
                line: node.loc.start.line,
                column: 0,
            };
            const startOfRange = sourceCode.getIndexFromLoc(start);
            return {
                range: [startOfRange, endOfRange],
                loc: { start, end },
                indentColumn: node.loc.start.column,
            };
        }
    },
});
