/* File : example.i */
%module example

%include "exception.i"

// The %exception directive will catch any exception thrown by the C++ library and
// panic() with the same message.
%exception {
    try {
        $action;
    } catch (std::exception &e) {
        _swig_gopanic(e.what());
    }
}

%{
#include <stdexcept>
#include "example.h"
%}

/* Let's just grab the original header file here */
%include "example.h"
