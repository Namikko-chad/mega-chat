import 'package:flutter/material.dart';

import '../../core/theme/colors.dart';
import '../../core/theme/typography.dart';

/// A round call control with its label under it: mute, camera, hang up.
class HmCallButton extends StatelessWidget {
  const HmCallButton({super.key, required this.icon, required this.label, required this.onPressed, this.color, this.active = false, this.size = 56});

  final IconData icon;
  final String label;
  final VoidCallback? onPressed;

  /// Filled with this colour (accept, hang up); otherwise a quiet button
  /// that lights up when [active].
  final Color? color;
  final bool active;
  final double size;

  @override
  Widget build(BuildContext context) {
    final fill = color ?? (active ? HmColors.text : HmColors.tint(HmColors.text, HmColors.badge));
    final fg = color != null ? HmColors.text : (active ? HmColors.bg : HmColors.text);
    return Semantics(
      button: true,
      label: label,
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Material(
            type: MaterialType.circle,
            color: fill,
            child: InkWell(
              customBorder: const CircleBorder(),
              onTap: onPressed,
              child: SizedBox.square(
                dimension: size,
                child: Icon(icon, color: fg, size: size * 0.45),
              ),
            ),
          ),
          const SizedBox(height: 6),
          Text(label, style: HmText.footer, textAlign: TextAlign.center),
        ],
      ),
    );
  }
}
