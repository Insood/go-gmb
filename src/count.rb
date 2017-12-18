counter = Hash.new(0)

File.open("out.txt").each do |line|
	opcode = line.gsub("\x0", "")[7..8]
	next if opcode == "in"
	counter [opcode] += 1
end

values = counter.collect { |k,v| [k,v] }.sort{ |this, that| this[1] <=> that [1] }
values.each do |k,v|
	print "#{k} : #{v}\n"
end