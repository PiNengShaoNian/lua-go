package state

import "testing"

func TestStack(t *testing.T) {
	ls := New()

	var str string

	ls.PushBoolean(true)
	str = stringifyStack(ls)
	if str != `[true]` {
		t.Fatalf("%s != [true]", str)
	}

	ls.PushInteger(10)
	str = stringifyStack(ls)
	if str != `[true][10]` {
		t.Fatalf("%s != [true][10]", str)
	}

	ls.PushNil()
	str = stringifyStack(ls)
	if str != `[true][10][nil]` {
		t.Fatalf("%s != [true][10][nil]", str)
	}

	ls.PushString("hello")
	str = stringifyStack(ls)
	if str != `[true][10][nil]["hello"]` {
		t.Fatalf(`%s != [true][10][nil]["hello"]`, str)
	}

	ls.PushValue(-4)
	str = stringifyStack(ls)
	if str != `[true][10][nil]["hello"][true]` {
		t.Fatalf(`%s != [true][10][nil]["hello"][true]`, str)
	}

	ls.Replace(3)
	str = stringifyStack(ls)
	if str != `[true][10][true]["hello"]` {
		t.Fatalf(`%s != [true][10][true]["hello"]`, str)
	}

	ls.SetTop(6)
	str = stringifyStack(ls)
	if str != `[true][10][true]["hello"][nil][nil]` {
		t.Fatalf(`%s != [true][10][true]["hello"][nil][nil]`, str)
	}

	ls.Remove(-3)
	str = stringifyStack(ls)
	if str != `[true][10][true][nil][nil]` {
		t.Fatalf(`%s != [true][10][true][nil][nil]`, str)
	}

	ls.SetTop(-5)
	str = stringifyStack(ls)
	if str != `[true]` {
		t.Fatalf(`%s != [true]`, str)
	}
}
