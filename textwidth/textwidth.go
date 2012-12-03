//Package textwidth provides a function to give the size of given text in given font in px.
package textwidth

/*
#include<stdio.h>
#include<stdlib.h>
#include<string.h>
#include<X11/Xlib.h>

typedef struct _Fnt {
	XFontStruct *xfont;
	XFontSet set;
	int ascent;
	int descent;
	int height;
} Fnt;

Fnt font;
Display *dpy;

unsigned int
textw(const char *text, unsigned int len) {
	XRectangle r;

	if(font.set) {
		XmbTextExtents(font.set, text, len, NULL, &r);
		return r.width;
	}
	return XTextWidth(font.xfont, text, len);
}

void
setfont(const char *fontstr) {
	char *def, **missing;
	int i, n;

	missing = NULL;
	if(font.set)
		XFreeFontSet(dpy, font.set);
	font.set = XCreateFontSet(dpy, fontstr, &missing, &n, &def);
	if(missing)
		XFreeStringList(missing);
	if(font.set) {
		//XFontSetExtents *font_extents;
		XFontStruct **xfonts;
		char **font_names;
		font.ascent = font.descent = 0;
		//font_extents = XExtentsOfFontSet(font.set);
		n = XFontsOfFontSet(font.set, &xfonts, &font_names);
		for(i = 0, font.ascent = 0, font.descent = 0; i < n; i++) {
			if(font.ascent < (*xfonts)->ascent)
				font.ascent = (*xfonts)->ascent;
			if(font.descent < (*xfonts)->descent)
				font.descent = (*xfonts)->descent;
			xfonts++;
		}
	}
	else {
		if(font.xfont)
			XFreeFont(dpy, font.xfont);
		font.xfont = NULL;
		if(!(font.xfont = XLoadQueryFont(dpy, fontstr))) {
			fprintf(stderr, "error, cannot load font: '%s'\n", fontstr);
			exit(EXIT_FAILURE);
		}
		font.ascent = font.xfont->ascent;
		font.descent = font.xfont->descent;
	}
	font.height = font.ascent + font.descent;
}
int
prepare()
{

	dpy = XOpenDisplay(0);
	if(!dpy) {
		fprintf(stderr, "cannot open display\n");
		return EXIT_FAILURE;
	}

	//setfont(myfont);
	//printf("%u\n", textw(text, strlen(text)));

	return EXIT_SUCCESS;
}
*/
//#cgo LDFLAGS: -lX11
import "C"

//Get takes the fully qualified font name of a font and some text, and returns an int64 of the text's length.
func Get(fontName, text string) int64 {
	C.setfont(C.CString(fontName))
	return int64(C.textw(C.CString(text), C.uint(len(text)+1)))
}
func init() {
	C.prepare()
}
