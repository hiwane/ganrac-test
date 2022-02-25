%{
package ganrac
%}

%union{
	node pNode
	num int
}

%token call list dict initvar f_time
%token name ident vardol varhist number f_true f_false t_str
%token all ex and or not abs
%token plus minus comma mult div pow
%token ltop gtop leop geop neop eqop assign
%token eol eolq lb rb lp rp lc rc

%type <num> seq_dict_arg seq_mobj list_mobj seq_ident
%type <node> f_true f_false eol eolq
%type <node> mobj lb initvar f_time
%type <node> vardol varhist number name ident t_str
%type <node> plus minus mult div pow and or
%type <node> ltop gtop leop geop neop eqop assign lb lp lc

%right assign
%left or
%left and
%left ltop gtop leop geop neop eqop
%left plus minus
%left mult div
%left unaryminus unaryplus
%right pow

%%


expr
	: eol                          { yylex.(*pLexer).push(newPNode("", eolq, 0, $1.pos))}
	| eolq                         { yylex.(*pLexer).push(newPNode("", eolq, 0, $1.pos))}
	| mobj eol                     { yylex.(*pLexer).push(newPNode("", eol, 0, $2.pos)) }
	| mobj eolq                    { yylex.(*pLexer).push(newPNode("", eolq, 1, $2.pos)) }
	;

mobj
	: number	{ yylex.(*pLexer).trace("int:" + $1.str); yylex.(*pLexer).push($1) }
	| t_str { yylex.(*pLexer).trace("string"); yylex.(*pLexer).push($1) }
	| f_true  { yylex.(*pLexer).trace("true");  yylex.(*pLexer).push($1)}
	| f_false { yylex.(*pLexer).trace("false"); yylex.(*pLexer).push($1)}
	| ident		{ yylex.(*pLexer).trace("ident: " + $1.str ); yylex.(*pLexer).push($1) }
	| name		{ yylex.(*pLexer).trace("name: " + $1.str ); yylex.(*pLexer).push($1) }
	| vardol		{ yylex.(*pLexer).trace("vardol: " + $1.str ); yylex.(*pLexer).push($1) }
	| varhist		{ yylex.(*pLexer).trace("varhist: " + $1.str ); yylex.(*pLexer).push($1) }
	| mobj and mobj { yylex.(*pLexer).trace("and"); yylex.(*pLexer).push($2)}
	| mobj or mobj  { yylex.(*pLexer).trace("or");  yylex.(*pLexer).push($2)}
	| lp mobj rp { $$ = $2 }
	| ident lp seq_mobj rp { yylex.(*pLexer).trace("call"); yylex.(*pLexer).push(newPNode($1.str, call, $3, $1.pos)) }
	| f_time lp mobj rp { yylex.(*pLexer).trace("time"); yylex.(*pLexer).push(newPNode($1.str, f_time, f_time, $1.pos)) }
	| ident lp rp { yylex.(*pLexer).trace("call"); yylex.(*pLexer).push(newPNode($1.str, call, 0, $1.pos)) }
	| mobj plus mobj	{ yylex.(*pLexer).trace("+"); yylex.(*pLexer).push($2)}
	| mobj minus mobj	{ yylex.(*pLexer).trace("-"); yylex.(*pLexer).push($2)}
	| mobj mult mobj	{ yylex.(*pLexer).trace("*"); yylex.(*pLexer).push($2)}
	| mobj div mobj		{ yylex.(*pLexer).trace("/"); yylex.(*pLexer).push($2)}
	| mobj pow mobj		{ yylex.(*pLexer).trace("^"); yylex.(*pLexer).push($2)}
	| minus mobj %prec unaryminus	{ yylex.(*pLexer).trace("-"); yylex.(*pLexer).push(newPNode("-.", unaryminus, 0, $1.pos)) }
	| plus mobj %prec unaryplus	{ yylex.(*pLexer).trace("+."); }
	| mobj ltop mobj { yylex.(*pLexer).trace("<");  yylex.(*pLexer).push($2)}
	| mobj gtop mobj { yylex.(*pLexer).trace(">");  yylex.(*pLexer).push($2)}
	| mobj leop mobj { yylex.(*pLexer).trace("<="); yylex.(*pLexer).push($2)}
	| mobj geop mobj { yylex.(*pLexer).trace(">="); yylex.(*pLexer).push($2)}
	| mobj eqop mobj { yylex.(*pLexer).trace("=="); yylex.(*pLexer).push($2)}
	| mobj neop mobj { yylex.(*pLexer).trace("!="); yylex.(*pLexer).push($2)}
	| list_mobj {}
	| dict_mobj { yylex.(*pLexer).trace("dict")}
	| mobj lb mobj rb { yylex.(*pLexer).trace("[]"); yylex.(*pLexer).push(newPNode("[]", lb, 0, $1.pos)) }
	| mobj assign mobj { yylex.(*pLexer).trace("=");  yylex.(*pLexer).push(newPNode("=", assign,   0, $1.pos))}
	| initvar lp seq_ident rp { yylex.(*pLexer).push(newPNode($1.str, initvar, $3, $1.pos)); }
	| initvar lp rp           { yylex.(*pLexer).push(newPNode($1.str, initvar,  0, $1.pos)); }
	;

// 辞書 {a: 1, b: x^2, c: "gao"}
dict_mobj
	: lc rc {
			yylex.(*pLexer).trace("dict0");
			yylex.(*pLexer).push(newPNode("_dict", dict, 0, $1.pos))
		}
	| lc seq_dict_arg rc {
			yylex.(*pLexer).trace("dictn" + string($2));
			yylex.(*pLexer).push(newPNode("_dict", dict, $2, $1.pos))
		}
	;

seq_dict_arg
	: ident eolq mobj {
			yylex.(*pLexer).trace("seqdi1:" + $1.str);
			$$ = 1;
			yylex.(*pLexer).push($1)
		}
	| t_str eolq mobj {
			yylex.(*pLexer).trace("seqds1:" + $1.str);
			$$ = 1;
			yylex.(*pLexer).push($1)
		}
	| seq_dict_arg comma ident eolq mobj {
			yylex.(*pLexer).trace("seqdin");
			$$ = $1 + 1;
			yylex.(*pLexer).push($3)
		}
	| seq_dict_arg comma t_str eolq mobj {
			yylex.(*pLexer).trace("seqdsn");
			$$ = $1 + 1;
			yylex.(*pLexer).push($3)
		}
	;

list_mobj
	: lb seq_mobj rb { yylex.(*pLexer).trace("list" + string($2)); yylex.(*pLexer).push(newPNode("_list", list, $2, $1.pos)) }
	| lb rb { yylex.(*pLexer).trace("list0"); yylex.(*pLexer).push(newPNode("_list", list, 0, $1.pos)) }
	;

seq_mobj
	: mobj	{ $$ = 1 }
	| seq_mobj comma mobj { $$ = $1 + 1 }
	;

seq_ident
	: ident	{ $$ = 1; yylex.(*pLexer).push(newPNode($1.str, ident, 0, $1.pos)) }
	| seq_ident comma ident { $$ = $1 + 1; yylex.(*pLexer).push(newPNode($3.str, ident, 0, $3.pos)) }
	;


%%      /*  start  of  programs  */
