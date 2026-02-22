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
    source_file: $ => repeat($._expression),

    _expression: $ => choice(
      $._expression_without_block,
      $._expression_with_block,
    ),

    _expression_without_block: $ => choice(
      $.binary_expression,
      $.unary_expression,
      $.number,
      $.boolean,
      $.identifier,
      $.parenthesized_expression,
      $.function_call,
      $.assignment,
      $.variable_declaration,
      $.function_declaration,
    ),

    _expression_with_block: $ => choice(
      $.block,
      $.if_expression,
    ),

    comment: $ => token(choice(
      seq('//', /.*/),
      seq('/*', /[^*]*\*+([^/*][^*]*\*+)*/, '/')
    )),

    variable_declaration: $ => seq(
      field('type', $.type),
      field('name', $.identifier),
      '=',
      field('value', $._expression),
    ),

    assignment: $ => seq(
      field('left', $.identifier),
      '=',
      field('right', $._expression),
    ),

    if_expression: $ => seq(
      'if',
      field('condition', $._expression),
      field('consequence', $.block),
      optional(seq(
        'else',
        field('alternative', choice($.block, $.if_expression))
      ))
    ),

    block: $ => seq(
      '{',
      repeat($._expression),
      '}'
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

    identifier: $ => /[a-zA-Z_][a-zA-Z0-9_]*/,

    number: $ => token(choice(
      seq(
        choice(
          /[0-9]+\.[0-9]+([eE][+-]?[0-9]+)?/,
          /[0-9]+/,
          /0[xX][0-9a-fA-F]+/,
          /0[bB][01]+/
        )
      )
    )),

    boolean: $ => choice('true', 'false'),
  }
});
