/* File : class.cxx */

#include <stdexcept>

#include "example.h"

double Line::perimeter() {
  return length;
}

double Line::area() {
  throw std::runtime_error("line does not have an area");
  return -1;
}
