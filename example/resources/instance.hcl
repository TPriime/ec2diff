resource "aws_instance" "example" {
  ami                    = "ami-0c94855ba95c71c99"
  instance_type          = "t2.micro"
  vpc_security_group_ids = ["sg-0123456789abcdef0"]
  key_name               = "your-key-name"
  security_groups        = ["your-security-group"]
  tags = {
    Name = "example-instance"
    Env  = "dev"
  }
  instance_state         = "running"
  public_ip              = "your-public-ip"
}