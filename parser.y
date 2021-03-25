%{
package ganrac
%}

%union{
	node pNode
	num int
}

%token call list initvar
%token name ident number f_true f_false t_str
%token all ex and or not abs
%token plus minus comma mult div pow
%token ltop gtop leop geop neop eqop assign
%token eol lb rb lp rp lc rc

%type <num> seq_mobj list_mobj seq_ident
%type <node> f_true f_false eol
%type <node> mobj lb initvar
%type <node> number name ident t_str
%type <node> plus minus mult div pow and or
%type <node> ltop gtop leop geop neop eqop assign lb lp

%left or
%left and
%left ltop gtop leop geop neop eqop
%left plus minus
%left mult div
%left unaryminus unaryplus
%right pow

%%

expr
	: eol { stack.Push(newPNode("", eol, 0, $1.pos))}
	| mobj eol { yyytrace("gege") }
	| name assign mobj eol  { yyytrace("assign"); stack.Push(newPNode($1.str, assign, 0, $1.pos)) }
	| initvar lp seq_ident rp eol { yyytrace("init"); stack.Push(newPNode($1.str, initvar, $3, $1.pos)) }
	;

mobj
	: number	{ yyytrace("poly.num: " + $1.str); stack.Push($1) }
	| t_str { yyytrace("string"); stack.Push($1) }
	| f_true  { yyytrace("true");  stack.Push($1)}
	| f_false { yyytrace("false"); stack.Push($1)}
	| ident		{ yyytrace("ident: " + $1.str ); stack.Push($1) }
	| name		{ yyytrace("name: " + $1.str ); stack.Push($1) }
	| mobj and mobj { yyytrace("and"); stack.Push($2)}
	| mobj or mobj  { yyytrace("or");  stack.Push($2)}
	| lp mobj rp { $$ = $2 }
	| ident lp seq_mobj rp { yyytrace("call"); stack.Push(newPNode($1.str, call, $3, $1.pos)) }
	| ident lp rp { yyytrace("call"); stack.Push(newPNode($1.str, call, 0, $1.pos)) }
	| mobj plus mobj	{ yyytrace("+"); stack.Push($2)}
	| mobj minus mobj	{ yyytrace("-"); stack.Push($2)}
	| mobj mult mobj	{ yyytrace("*"); stack.Push($2)}
	| mobj div mobj		{ yyytrace("/"); stack.Push($2)}
	| mobj pow mobj		{ yyytrace("^"); stack.Push($2)}
	| minus mobj %prec unaryminus	{ yyytrace("-"); stack.Push(newPNode("-.", unaryminus, 0, $1.pos)) }
	| plus mobj %prec unaryplus	{ yyytrace("+."); }
	| mobj ltop mobj { yyytrace("<");  stack.Push($2)}
	| mobj gtop mobj { yyytrace(">");  stack.Push($2)}
	| mobj leop mobj { yyytrace("<="); stack.Push($2)}
	| mobj geop mobj { yyytrace(">="); stack.Push($2)}
	| mobj eqop mobj { yyytrace("=="); stack.Push($2)}
	| mobj neop mobj { yyytrace("!="); stack.Push($2)}
	| list_mobj {}
	| mobj lb mobj rb { yyytrace("="); stack.Push(newPNode("[]", lb, 0, $1.pos)) }
	;

list_mobj
	: lb seq_mobj rb { yyytrace("list" + string($2)); stack.Push(newPNode("_list", list, $2, $1.pos)) }
	| lb rb { yyytrace("list0"); stack.Push(newPNode("_list", list, 0, $1.pos)) }
	;

seq_mobj
	: mobj	{ $$ = 1 }
	| seq_mobj comma mobj { $$ = $1 + 1 }
	;

seq_ident
	: ident	{ $$ = 1; stack.Push(newPNode($1.str, ident, 0, $1.pos)) }
	| seq_ident comma ident { $$ = $1 + 1; stack.Push(newPNode($3.str, ident, 0, $3.pos)) }
	;


%%      /*  start  of  programs  */
