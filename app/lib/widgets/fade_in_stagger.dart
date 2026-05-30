import 'package:flutter/material.dart';

/// Staggered fade-in animation widget for sequential entry effects.
class FadeInStagger extends StatefulWidget {
  final Widget child;
  final int delayMs;

  const FadeInStagger({
    super.key,
    required this.child,
    required this.delayMs,
  });

  @override
  State<FadeInStagger> createState() => _FadeInStaggerState();
}

class _FadeInStaggerState extends State<FadeInStagger> {
  double _opacity = 0.0;

  @override
  void initState() {
    super.initState();
    Future.delayed(Duration(milliseconds: widget.delayMs), () {
      if (mounted) {
        setState(() => _opacity = 1.0);
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedOpacity(
      opacity: _opacity,
      duration: const Duration(milliseconds: 300),
      child: widget.child,
    );
  }
}
