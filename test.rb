regs = ["B","C","D","E","H", "L", "(HL)","A"]

(64..128).each do |i|
	register = i & 0x7
	bit      = (i >> 3) & 0x7
	print sprintf("%X BIT #{bit} #{regs[register]}\n",i)
end