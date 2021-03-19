%{
package ganrac
%}

%union{
	node pNode
	num int
}

%token call list initvar
%token name ident number f_true f_false
%token all ex and or not abs
%token plus minus comma mult div pow
%token ltop gtop leop geop neop eqop assign
%token eol lb rb lp rp lc rc

%type <num> seq_mobj list_mobj seq_ident
%type <node> f_true f_false
%type <node> fof atom mobj lb initvar
%type <node> number poly name ident
%type <node> plus minus mult div pow and or
%type <node> ltop gtop leop geop neop eqop assign lb

%left or
%left and
%left ltop gtop leop geop neop eqop
%left plus minus
%left mult div
%left unaryminus unaryplus
%right pow

%%

expr
	: mobj eol {{ yyytrace("gege") }}
	| name assign mobj eol  { yyytrace("assign"); stack.Push($2); }
	;

mobj
	: fof
	| poly {{ yyytrace("mobj: poly:" + string(stack.Len())) }}
	| ident lp seq_ident rp { yyytrace("call"); stack.Push(newPNode($1.str, call, $3, $1.pos)) }
	| initvar lp seq_mobj rp { yyytrace("init"); stack.Push(newPNode($1.str, initvar, $3, $1.pos)) }
	| list_mobj {}
	;

fof
	: atom
	| fof and fof { yyytrace("and"); stack.Push($2)}
	| fof or fof  { yyytrace("or");  stack.Push($2)}
	| lp fof rp { $$ = $2 }
	| name		{ yyytrace("name: " + $1.str ); stack.Push($1) }
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
	: ident	{ $$ = 1 }
	| seq_ident comma ident { $$ = $1 + 1 }
	;

atom
	: f_true  { yyytrace("true");  stack.Push($1)}
	| f_false { yyytrace("false"); stack.Push($1)}
	| poly ltop poly { yyytrace("<");  stack.Push($2)}
	| poly gtop poly { yyytrace(">");  stack.Push($2)}
	| poly leop poly { yyytrace("<="); stack.Push($2)}
	| poly geop poly { yyytrace(">="); stack.Push($2)}
	| poly eqop poly { yyytrace("=="); stack.Push($2)}
	| poly neop poly { yyytrace("!="); stack.Push($2)}
	;

/*
rational
	: number	{ yyytrace("rat.num: " + $1.str); stack.Push($1) }
	| lp rational rp	{ $$ = $2 }
	| rational plus rational	{ yyytrace("+"); stack.Push($2)}
	| rational minus rational	{ yyytrace("-"); stack.Push($2)}
	| rational mult rational	{ yyytrace("*"); stack.Push($2)}
	| rational div rational	{ yyytrace("/"); stack.Push($2)}
	| minus rational %prec unaryminus	{ yyytrace("-"); newPNode("-.", unaryminus, 0, $1.pos) }
	| plus rational %prec unaryplus	{ yyytrace("+"); newPNode("+.", unaryplus, 0, $1.pos) }
	;
*/

poly
	: lp poly rp { $$ = $2; }
	| number	{ yyytrace("poly.num: " + $1.str); stack.Push($1) }
	| ident		{ yyytrace("ident: " + $1.str ); stack.Push($1) }
	| poly plus poly	{ yyytrace("+"); stack.Push($2)}
	| poly minus poly	{ yyytrace("-"); stack.Push($2)}
	| poly mult poly	{ yyytrace("*"); stack.Push($2)}
	| poly div poly		{ yyytrace("/"); stack.Push($2)}
	| poly pow poly		{ yyytrace("^"); stack.Push($2)}
	| minus poly %prec unaryminus	{ yyytrace("-"); stack.Push(newPNode("-.", unaryminus, 0, $1.pos)) }
	| plus poly %prec unaryplus	{ yyytrace("+"); stack.Push(newPNode("+.", unaryplus, 0, $1.pos)) }
	;

%%      /*  start  of  programs  */
