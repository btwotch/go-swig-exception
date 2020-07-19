/* File : example.h */

class Line {
private:
  double length;
public:
  Line(double l) : length(l) { }
  double area();
  double perimeter();
};
