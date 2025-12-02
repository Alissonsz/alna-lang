module.exports = grammar({
  name: 'alna',

  extras: $ => [
    /\s/,
    $.comment,
  ],

  conflicts: $ => [
    [$._expression, $.function_call],
  ],

  rules: {
    source_file: $ => repeat($._statement),

    _statement: $ => choice(
      $.variable_declaration,
      $.assignment,
      $.if_statement,
      $.while_statement,
      $.for_statement,
      $.return_statement,
      $.break_statement,
      $.continue_statement,
      $.expression_statement,
      $.block,
      $.function_declaration,
    ),

    // Comments
    comment: $ => token(choice(
      seq('//', /.*/),
      seq('/*', /[^*]*\*+([^/*][^*]*\*+)*/, '/')
    )),

    // Variable declaration: int x = 10
    variable_declaration: $ => seq(
      field('type', $.type),
      field('name', $.identifier),
      '=',
      field('value', $._expression),
    ),

    // Assignment: x = 5
    assignment: $ => seq(
      field('left', $.identifier),
      '=',
      field('right', $._expression),
    ),

    // If statement
    if_statement: $ => seq(
      'if',
      field('condition', $._expression),
      field('consequence', $.block),
      optional(seq(
        'else',
        field('alternative', choice($.block, $.if_statement))
      ))
    ),

    // While statement
    while_statement: $ => seq(
      'while',
      field('condition', $._expression),
      field('body', $.block)
    ),

    // For statement
    for_statement: $ => seq(
      'for',
      field('initializer', optional($._statement)),
      ';',
      field('condition', optional($._expression)),
      ';',
      field('update', optional($._expression)),
      field('body', $.block)
    ),

    // Return statement
    return_statement: $ => prec.left(seq(
      'return',
      optional($._expression)
    )),

    // Break statement
    break_statement: $ => 'break',

    // Continue statement
    continue_statement: $ => 'continue',

    // Block: { ... }
    block: $ => seq(
      '{',
      repeat($._statement),
      '}'
    ),

    // Expression statement
    expression_statement: $ => $._expression,

    // Expressions
    _expression: $ => choice(
      $.binary_expression,
      $.unary_expression,
      $.number,
      $.boolean,
      $.identifier,
      $.parenthesized_expression,
      $.function_call,
    ),

    binary_expression: $ => {
      const table = [
        [7, choice('*', '/', '%')],
        [6, choice('+', '-')],
        [5, choice('<', '<=', '>', '>=')],
        [4, choice('==', '!=')],
        [3, '&&'],
        [2, '||'],
      ];

      return choice(...table.map(([precedence, operator]) =>
        prec.left(precedence, seq(
          field('left', $._expression),
          field('operator', operator),
          field('right', $._expression)
        ))
      ));
    },

    unary_expression: $ => prec(8, seq(
      field('operator', choice('!', '-')),
      field('operand', $._expression)
    )),

    parenthesized_expression: $ => seq(
      '(',
      $._expression,
      ')'
    ),

    function_call: $ => seq(
      field('function', $.identifier),
      '(',
      optional($.argument_list),
      ')'
    ),

    argument_list: $ => seq(
      $._expression,
      repeat(seq(',', $._expression))
    ),

    // Function declaration
    function_declaration: $ => seq(
      field('return_type', $.type),
      field('name', $.identifier),
      '(',
      optional($.parameter_list),
      ')',
      field('body', $.block)
    ),

    parameter_list: $ => seq(
      $.parameter,
      repeat(seq(',', $.parameter))
    ),

    parameter: $ => seq(
      field('type', $.type),
      field('name', $.identifier)
    ),

    // Types
    type: $ => choice(
      'int',
      'i8',
      'i16',
      'i32',
      'i64',
      'u8',
      'u16',
      'u32',
      'u64',
      'float',
      'f32',
      'f64',
      'bool',
      'string',
      'void'
    ),

    // Primitives
    identifier: $ => /[a-zA-Z_][a-zA-Z0-9_]*/,

    number: $ => token(choice(
      seq(
        choice(
          /[0-9]+\.[0-9]+([eE][+-]?[0-9]+)?/,  // float
          /[0-9]+/,                              // decimal
          /0[xX][0-9a-fA-F]+/,                   // hex
          /0[bB][01]+/                           // binary
        )
      )
    )),

    boolean: $ => choice('true', 'false'),
  }
});
