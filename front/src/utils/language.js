function getDefinition() {
    return {
        case_insensitive: true,
        keywords: {
            keyword: 'any and not is alert inc dec set',
            literal:
                'ingress contains icontains regex array json form path query body headers cookies',
        },
        contains: [
            {
                className: 'string',
                begin: "'",
                end: "'",
            },
        ],
    };
}

export { getDefinition };
